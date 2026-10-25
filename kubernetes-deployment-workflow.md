# Kubernetes Deployment Workflow: Go Microservices, Ingress, Load Balancing, Scaling, and Observability

This guide explains how the Go microservices project in `~/Desktop/micro/project` fits together when deployed to Kubernetes with Minikube.

The project includes services such as the broker, authentication, logger, listener, and mailer, plus supporting services such as MongoDB and RabbitMQ. The exact traffic paths depend on the Kubernetes YAML manifests, ports, selectors, and Ingress rules in the project.

## 1. Complete Architecture

```text
                         CLIENT / BROWSER
                                |
                                v
                         INGRESS CONTROLLER
                  (HTTP/HTTPS host/path routing)
                                |
                                v
                       KUBERNETES SERVICE
                   (stable name and virtual IP)
                                |
                                v
                      DEPLOYMENT / REPLICASET
                         manages replica count
                         /       |       \
                        v        v        v
                      POD 1    POD 2    POD 3
                        |        |        |
                     Go app   Go app   Go app
                    container container container

       INSIDE THE CLUSTER
       ------------------
       Broker -> other microservices and dependencies,
       according to application configuration.

       Supporting services may include MongoDB and RabbitMQ.

       OBSERVABILITY
       -------------
       Metrics -> Prometheus
       Traces  -> OpenTelemetry instrumentation/export -> Jaeger
       Logs    -> log collector/store (for example Fluent Bit
                  with Elasticsearch/Kibana)
```

This is a conceptual architecture. It does not mean every component above is already installed or connected.

## 2. Core Kubernetes Components

### 2.1 Container

A container runs your compiled Go application and its runtime dependencies. A Dockerfile defines how the image is built.

**Example:** the broker container runs the broker HTTP server.

### 2.2 Pod

A Pod is the smallest deployable unit in Kubernetes. It runs one or more containers and receives a cluster network identity. A replacement Pod can receive a different IP address.

**Example:** a Pod named something like `broker-service-xxxxx` runs the broker container.

### 2.3 Deployment

A Deployment declares how many replicas of an application should run and how updates should happen. Kubernetes uses a ReplicaSet to maintain the desired number of Pods.

For example, `replicas: 2` asks Kubernetes to keep two broker Pods running.

### 2.4 Service

A Service provides a stable name and virtual IP for a group of matching Pods. It forwards traffic to eligible endpoints, typically distributing connections across them.

**Example:** `broker-service` can keep the same DNS name while its Pods are replaced.

### 2.5 Ingress and Ingress Controller

An **Ingress** resource defines HTTP/HTTPS routing rules. An **Ingress Controller** implements those rules and forwards traffic to the appropriate Service.

For example, `front-end.info` could route to the front-end Service and `broker-service.info` to the broker Service if those hosts and backends are configured.

An Ingress resource alone does not guarantee a working route. The controller must be running, the backend Service and port must exist, and the hostname must resolve to a reachable address.

## 3. Load Balancing: Where Does It Happen?

There are several different mechanisms that people may call a Kubernetes load balancer.

### A. External Load Balancer

Distributes traffic from outside the cluster to a reachable Kubernetes entry point or Service. In cloud environments, a Service of type `LoadBalancer` commonly requests a cloud load balancer.

In local Minikube, an external IP may require a tunnel or another local access method.

### B. Ingress Controller

Routes HTTP/HTTPS traffic based on hostnames or URL paths. For example, a hostname could route to the front end while another routes to the broker.

An Ingress Controller can also handle TLS termination and other HTTP routing features.

### C. Kubernetes Service

Provides a stable address and forwards traffic to eligible Pod endpoints selected by labels and selectors. Kubernetes networking components implement this forwarding; it is not necessarily a separate load-balancer process.

### Example: Three Broker Replicas

```text
Client
  |
  v
Ingress Controller
  |  matches broker-service.info
  v
broker-service
  |  stable Service address
  +----------+----------+
  |          |          |
  v          v          v
 Pod 1      Pod 2      Pod 3
```

The Service forwards connections to eligible Pods, but this does not guarantee perfectly equal traffic per request.

**Important:** you do not need three different load balancers. External load balancing, Ingress routing, and Service forwarding have distinct roles. Not every deployment requires all three.

## 4. What Is a Kubernetes Cluster?

A cluster is the whole Kubernetes environment that manages your applications.

```text
KUBERNETES CLUSTER
|
+-- Control Plane
|   +-- API server
|   +-- Scheduler
|   +-- Controllers
|
+-- Worker Nodes
    +-- Application Pods
    +-- Containers
```

The control plane schedules and manages workloads. Worker nodes provide the environment where Pods run.

In this project, the command:

```bash
minikube start --nodes=2
```

configured a local Minikube cluster with a control-plane node and an additional node named `minikube-m02`.

This does **not** mean every microservice automatically gets two copies. Each Deployment's replica count controls the number of Pods.

Two nodes on one development machine are useful for learning, but they do not provide the same failure isolation as independent production servers.

## 5. Ingress in This Project

Commands previously used include:

```bash
minikube addons enable ingress
kubectl apply -f ingress.yml
kubectl get ingress
```

What they do:

- `minikube addons enable ingress` enables an Ingress Controller add-on in Minikube.
- `kubectl apply -f ingress.yml` creates or updates the routing rules in that file.
- `kubectl get ingress` lists Ingress resources and their configured hosts.

The earlier output showed an Ingress called `my-ingress` with hosts `front-end.info` and `broker-service.info`.

For local hostname routing to work, check that:

1. The Ingress Controller is running.
2. The referenced Services and ports exist.
3. The hostname resolves to a reachable address.
4. Your local Minikube networking setup permits access.

Depending on the setup, you may need entries in the Windows hosts file or another supported access method.

Useful inspection commands:

```bash
kubectl get ingress
kubectl describe ingress my-ingress
kubectl get pods -A
kubectl get svc
```

An Ingress resource by itself does not prove that the route is working.

## 6. Service Types: ClusterIP, NodePort, and LoadBalancer

| Service type | What it does | Typical use |
|---|---|---|
| `ClusterIP` | Exposes a Service inside the cluster | Broker, logger, database |
| `NodePort` | Exposes a port on cluster nodes | Development or specific external access setups |
| `LoadBalancer` | Requests an external load-balancing endpoint from the environment | Public-facing applications in supported environments |

A possible design for this project:

- A front-end Service reachable through Ingress.
- A broker Service reachable through Ingress if external access is intended.
- Internal `ClusterIP` Services for authentication, logger, listener, and mailer.
- Internal access for MongoDB and RabbitMQ unless there is a specific reason to expose them.

The actual configuration should follow the ports and selectors in the YAML. Avoid exposing databases or message brokers publicly without a clear security requirement.

If `kubectl get svc` shows `EXTERNAL-IP` as `<pending>` for a `LoadBalancer` Service, that is common when the environment does not automatically provide external load balancers. In Minikube, `minikube tunnel` can provide an external IP for supported `LoadBalancer` Services while the tunnel is running. It is not required for every Ingress setup.

## 7. Scaling: Where Does It Happen?

There are two main kinds of scaling.

### 7.1 Horizontal Pod Scaling

Adds or removes copies of an application.

```text
Deployment
    |
    +-- 1 Pod
    |
    +-- scale up -> 3 Pods -> 5 Pods
```

You can set a replica count manually or use a Horizontal Pod Autoscaler (HPA) to adjust replicas based on configured metrics.

Manually scale the broker Deployment:

```bash
kubectl scale deployment broker-service --replicas=3
kubectl get pods -o wide
kubectl get deployment broker-service
```

For autoscaling, you need an HPA configuration, suitable resource requests on containers, and a working metrics source. Metrics Server is commonly used for CPU-based scaling.

Check autoscaling and resource metrics:

```bash
kubectl get hpa
kubectl top nodes
kubectl top pods
```

If `kubectl top` reports that metrics are unavailable, check whether Metrics Server is installed and working.

### 7.2 Node Scaling

Adds or removes worker machines.

```text
Cluster: 2 nodes -> 3 nodes -> 4 nodes
```

In supported environments, a Cluster Autoscaler can adjust node capacity when Pods cannot be scheduled due to insufficient resources.

### How Load Balancing Relates to Scaling

When you add replicas, a Service can direct new connections to newly eligible Pods. The Service does not decide to create replicas; a Deployment or HPA manages that.

## 8. Observability: How Do You Know the System Is Healthy?

The Kubernetes dashboard opened with `minikube dashboard` helps inspect cluster resources. It is not, by itself, a complete observability stack.

### Metrics — Prometheus

Helps answer questions such as:

- How much CPU and memory are being used?
- How many requests fail?
- Is latency increasing?

### Distributed Traces — OpenTelemetry and Jaeger

Help follow a request across multiple services. For example, if a request enters the broker and calls another service, tracing can help identify which step is slow.

OpenTelemetry provides instrumentation and telemetry export; Jaeger can store and display traces.

### Logs — Log Collection and Storage

Logs help investigate failures and errors. A possible stack is Fluent Bit for collection and Elasticsearch/Kibana for storage and searching.

### Dashboards and Alerts

Dashboards bring signals together, while alerts can notify you when defined conditions occur.

Observability can help identify when scaling is needed. An HPA can consume supported metrics to scale Pods, but an observability dashboard does not scale them by itself.

## 9. Deployment Workflow and Commands

Run these commands from `~/Desktop/micro/project`.

### Step 1 — Start the Cluster

```bash
minikube start --nodes=2
kubectl get nodes
```

Both nodes should eventually report `Ready`.

### Step 2 — Apply Application Manifests

```bash
kubectl apply -f k8s/
```

This applies supported manifests in that directory. Check that required dependencies, configuration, secrets, and persistent storage are also set up.

### Step 3 — Check Deployments and Pods

```bash
kubectl get deployments
kubectl get pods -o wide
kubectl get events --sort-by=.metadata.creationTimestamp
```

Look for Pods in states such as `Pending`, `CrashLoopBackOff`, or `ImagePullBackOff`.

### Step 4 — Check Services and Ingress

```bash
kubectl get services
kubectl get ingress
kubectl describe ingress my-ingress
```

Verify Service selectors and ports, and ensure that the Ingress points to the intended Service and port.

### Step 5 — Test the Front End Locally

```bash
kubectl get svc front-end
kubectl port-forward svc/front-end 8080:<service-port>
```

Replace `<service-port>` with the actual Service port shown by `kubectl get svc front-end`. Then open:

`http://localhost:8080`

### Step 6 — Inspect a Failing Pod

```bash
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

If the Pod has multiple containers:

```bash
kubectl logs <pod-name> -c <container-name>
```

If it keeps restarting:

```bash
kubectl logs <pod-name> --previous
```

### Step 7 — Scale a Service

```bash
kubectl scale deployment broker-service --replicas=3
kubectl get pods -o wide
```

Check that the new Pods become ready and are included among the Service's endpoints.

### Important Note About Deleting a Service

The command:

```bash
kubectl delete svc broker-service
```

deletes the Service, not the Deployment or its Pods. Applying `k8s/broker.yml` afterward can recreate the Service if that manifest defines it.

## 10. The Complete Picture to Remember

```text
BUILD
  Docker packages each Go microservice into an image.
    |
    v
DEPLOY
  Deployments manage Pods running those images.
    |
    v
ROUTE
  Ingress handles HTTP routing.
  Services provide stable endpoints and forward traffic.
    |
    v
RUN
  Nodes provide compute resources.
  The control plane manages scheduling and desired state.
    |
    v
OBSERVE
  Metrics, traces, and logs help explain performance and failures.
    |
    v
SCALE
  Deployments or HPAs change Pod replicas.
  A Cluster Autoscaler can change node capacity when supported.
```

## Recommended Next Step

Inspect the actual manifests, especially:

- `k8s/front-end.yml`
- `k8s/broker.yml`
- `ingress.yml`

These files reveal which ports, Service types, selectors, hostnames, and replica counts your current project really uses.
