db = db.getSiblingDB('faaskubebench');

// Inserir funções de exemplo
if (db.functions.countDocuments() === 0) {
    db.functions.insertMany([
        {
            name: "matrix-multiplication",
            runtime: "python",
            code_path: "/functions/matrix-multiplication",
        },
        {
            name: "image-processing",
            runtime: "node",
            code_path: "/functions/image-processing",
        },
        {
            name: "hello-world",
            runtime: "go",
            code_path: "/functions/hello-world",
        },
        {
            name: "data-processing",
            runtime: "java",
            code_path: "/functions/data-processing.jar",
        }
    ]);
    print("Functions inserted");
}

// Inserir workloads de exemplo
if (db.workloads.countDocuments() === 0) {
    db.workloads.insertMany([
        {
            name: "cpu-intensive",
            description: "Workload intensivo em CPU",
            type: "cpu",
            functions: ["matrix-multiplication", "data-processing"],
        },
        {
            name: "memory-intensive",
            description: "Workload intensivo em memória", 
            type: "memory",
            functions: ["image-processing", "data-processing"],
        },
        {
            name: "basic-test",
            description: "Teste básico de função",
            type: "mixed",
            functions: ["hello-world"],
        },
        {
            name: "network-intensive",
            description: "Workload intensivo em rede",
            type: "network",
            functions: ["data-processing", "image-processing"],
        }
    ]);
    print("Workloads inserted");
}

// Inserir plataformas de exemplo
if (db.platforms.countDocuments() === 0) {
    db.platforms.insertMany([
        {
            name: "knative",
            type: "knative",
            api_endpoint: "",
            namespace: "default",
        },
        {
            name: "openfaas",
            type: "openfaas",
            api_endpoint: "",
            namespace: "default",
        },
        {
            name: "openwhisk",
            type: "openwhisk",
            api_endpoint: "",
            namespace: "default",
        },
    ]);
    print("Platforms inserted");
}

print("Database faaskubebench initialized successfully!");