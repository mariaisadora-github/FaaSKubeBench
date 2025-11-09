#!/usr/bin/env python3
import time
import os
from prometheus_client import start_http_server, Gauge
from kubernetes import client, config
from kubernetes.client.rest import ApiException
from datetime import datetime, timedelta

try:
    config.load_incluster_config()
except config.ConfigException:
    try:
        config.load_kube_config()
    except config.ConfigException:
        print("Não foi possível configurar o cliente Kubernetes. Verifique se está em um cluster ou se o kubeconfig está disponível.")
        exit(1)

v1 = client.CoreV1Api()
custom_objects_api = client.CustomObjectsApi()

# --- Métricas Prometheus ---
# Tempo de inicialização do pod (started_at - start_time)
BOOT_GAUGE = Gauge(
    'serverless_pod_boot_duration_seconds',
    'Tempo de inicialização do pod (started_at - start_time)',
    ['namespace', 'pod', 'function']
)

# Uso de CPU por pod (millicores) - do Metrics Server
CPU_GAUGE = Gauge(
    'serverless_pod_cpu_usage_millicores',
    'Uso de CPU por pod (millicores) do Metrics Server',
    ['namespace', 'pod', 'function']
)

# Uso de memória por pod (bytes) - do Metrics Server
MEM_GAUGE = Gauge(
    'serverless_pod_memory_usage_bytes',
    'Uso de memória por pod (bytes) do Metrics Server',
    ['namespace', 'pod', 'function']
)

# Número de pods ativos por plataforma e função
POD_COUNT_GAUGE = Gauge(
    'serverless_pod_count',
    'Número de pods ativos por plataforma e função',
    ['platform', 'function', 'namespace']
)

# Diferença na contagem de pods (pós-benchmark - pré-benchmark)
POD_SCALED_DIFF_GAUGE = Gauge(
    'serverless_pod_scaled_difference',
    'Diferença na contagem de pods (pós-benchmark - pré-benchmark)',
    ['platform', 'function', 'namespace']
)

# --- Identificadores por plataforma (Labels para identificar funções) ---
PLATFORM_LABELS = {
    'knative': 'serving.knative.dev/service',
    'openwhisk': 'whisk-managed',
    'openfaas': 'faas_function',
    'fission': 'fission-function-name'
}

# --- Variáveis para armazenar estado do benchmark ---
initial_pod_counts = {}
benchmark_start_time = None

# --- Funções de Coleta de Métricas ---
def get_metrics_from_metrics_server(namespace, pod_name):
    """Obtém métricas de CPU e memória de um pod do Metrics Server."""
    try:
        # A API do Metrics Server é um Custom Object
        metrics = custom_objects_api.get_namespaced_custom_object(
            group='metrics.k8s.io',
            version='v1beta1',
            name=pod_name,
            namespace=namespace,
            plural='pods'
        )

        cpu_usage = 0
        memory_usage = 0

        for container in metrics['containers']:
            # CPU em millicores (n = nanocores / 1_000_000)
            cpu_raw = container['usage']['cpu']
            if cpu_raw.endswith('n'):
                cpu_usage += int(cpu_raw[:-1]) / 1_000_000
            elif cpu_raw.endswith('m'):
                cpu_usage += int(cpu_raw[:-1])
            else:
                cpu_usage += int(cpu_raw) / 1_000_000 # Assume que é em nanocores se não tiver sufixo

            # Memória em bytes (Ki = Kibibytes, Mi = Mebibytes, Gi = Gibibytes)
            mem_raw = container['usage']['memory']
            if mem_raw.endswith('Ki'):
                memory_usage += int(mem_raw[:-2]) * 1024
            elif mem_raw.endswith('Mi'):
                memory_usage += int(mem_raw[:-2]) * 1024 * 1024
            elif mem_raw.endswith('Gi'):
                memory_usage += int(mem_raw[:-2]) * 1024 * 1024 * 1024
            else:
                memory_usage += int(mem_raw) # Assume que é em bytes se não tiver sufixo

        return cpu_usage, memory_usage

    except ApiException as e:
        if e.status == 404:
            # print(f"Métricas para o pod {pod_name} no namespace {namespace} não encontradas (ainda ou Metrics Server não disponível).")
            pass # Pod pode ainda não ter métricas
        else:
            print(f"Erro ao obter métricas do Metrics Server para {pod_name}: {e}")
        return 0, 0
    except Exception as e:
        print(f"Erro inesperado ao processar métricas do pod {pod_name}: {e}")
        return 0, 0

def get_function_name_from_pod(pod_labels, platform_config):
    """Tenta identificar o nome da função a partir dos labels do pod."""
    for platform, label_key in platform_config.items():
        if label_key in pod_labels:
            return platform, pod_labels[label_key]
    return None, None

def collect_metrics_loop():
    """Coleta métricas de pods, CPU, memória e tempo de inicialização."""
    global initial_pod_counts

    current_pod_counts = {platform: {} for platform in PLATFORM_LABELS} # {platform: {function_name: count}}
    all_pods = []

    try:
        pods_list = v1.list_pod_for_all_namespaces(watch=False)
        all_pods = pods_list.items
    except ApiException as e:
        print(f"Erro ao listar pods: {e}")
        return

    for pod in all_pods:
        ns = pod.metadata.namespace
        name = pod.metadata.name
        labels = pod.metadata.labels or {}

        platform, function_name = get_function_name_from_pod(labels, PLATFORM_LABELS)
        if not platform or not function_name:
            continue  # Ignora pods que não são de funções serverless conhecidas

        # Atualiza contagem de pods por função e plataforma
        if function_name not in current_pod_counts[platform]:
            current_pod_counts[platform][function_name] = 0
        current_pod_counts[platform][function_name] += 1

        # Coleta CPU e Memória do Metrics Server
        cpu_usage, memory_usage = get_metrics_from_metrics_server(ns, name)
        if cpu_usage > 0:
            CPU_GAUGE.labels(namespace=ns, pod=name, function=function_name).set(cpu_usage)
        if memory_usage > 0:
            MEM_GAUGE.labels(namespace=ns, pod=name, function=function_name).set(memory_usage)

        # Tempo de inicialização do pod (started_at - start_time)
        try:
            if pod.status and pod.status.start_time:
                pod_start_time = pod.status.start_time.astimezone(time.timezone) # Garante timezone aware
                boot_time_set = False
                for container_status in pod.status.container_statuses or []:
                    if container_status.state and container_status.state.running and container_status.state.running.started_at:
                        container_started_at = container_status.state.running.started_at.astimezone(time.timezone)
                        boot_duration = (container_started_at - pod_start_time).total_seconds()
                        BOOT_GAUGE.labels(namespace=ns, pod=name, function=function_name).set(boot_duration)
                        boot_time_set = True
                        break
                if not boot_time_set:
                    # Se não encontrou container running, mas o pod já iniciou, pode ser um pod pendente ou em fase inicial
                    # Ou se o pod iniciou mas o container ainda não está 'running'
                    pass

        except Exception as e:
            # print(f"Erro ao calcular tempo de inicialização para pod {name}: {e}")
            pass # Ignora erro e não seta a métrica

    # Atualiza métricas de contagem de pods por plataforma e função
    for platform, functions in current_pod_counts.items():
        for function_name, count in functions.items():
            POD_COUNT_GAUGE.labels(platform=platform, function=function_name, namespace=ns).set(count)

            # Calcula diferença de pods se o benchmark já começou
            if benchmark_start_time and platform in initial_pod_counts and function_name in initial_pod_counts[platform]:
                diff = count - initial_pod_counts[platform][function_name]
                POD_SCALED_DIFF_GAUGE.labels(platform=platform, function=function_name, namespace=ns).set(diff)


def set_initial_pod_counts():
    global initial_pod_counts
    print("Capturando contagem inicial de pods...")
    pods_list = v1.list_pod_for_all_namespaces(watch=False)
    for pod in pods_list.items:
        ns = pod.metadata.namespace
        name = pod.metadata.name
        labels = pod.metadata.labels or {}

        platform, function_name = get_function_name_from_pod(labels, PLATFORM_LABELS)
        if platform and function_name:
            if platform not in initial_pod_counts:
                initial_pod_counts[platform] = {}
            if function_name not in initial_pod_counts[platform]:
                initial_pod_counts[platform][function_name] = 0
            initial_pod_counts[platform][function_name] += 1
    print(f"Contagem inicial de pods capturada: {initial_pod_counts}")

def main():
    global benchmark_start_time

    start_http_server(8000)
    print("Exporter Prometheus rodando na porta 8000")

    set_initial_pod_counts()

    benchmark_start_time = datetime.now(time.timezone) 
    print(f"Benchmark simulado iniciado em: {benchmark_start_time}")

    while True:
        collect_metrics_loop()
        time.sleep(os.getenv('COLLECTION_INTERVAL_SECONDS', 10))

if __name__ == '__main__':
    main()
