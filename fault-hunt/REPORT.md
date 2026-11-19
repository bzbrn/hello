# Fault hunt report

Fill in one section per fault, in the order you found them. Paste real output, not a summary.

## Fault 1

- **Symptom** (what you saw first): Pods stuck in Pending with 0/1
- **Command** (that showed the cause): kubectl describe pod hello-7f6755ddc7-77df8
- **Evidence** (the exact output line): Warning  FailedScheduling  6m54s (x2 over 11m)  default-scheduler  0/1 nodes are available: 1 Insufficient cpu. preemption: 0/1 nodes are available: 1 Preemption is not helpful for scheduling
- **Cause**: It asks for 64 CPU cores which are way too many
- **Fix** (file and field you changed): the line `cpu: "64"` in deployment.yaml changed to `cpu: 50m`

## Fault 2

- **Symptom** (what you saw first): Newest pod stuck in ImagePullBackOff
- **Command** (that showed the cause): kubectl describe pod hello-77c576fd4d-ccxzx
- **Evidence** (the exact output line): Warning  Failed     3m40s (x5 over 6m30s)  kubelet            spec.containers{hello}: Failed to pull image "hello:0.1.10": failed to pull and unpack image "docker.io/library/hello:0.1.10": failed to resolve reference "docker.io/library/hello:0.1.10": pull access denied, repository does not exist or may require authorization: server message: insufficient_scope: authorization failed 
- **Cause**: There is no 0.1.10 tag, only 0.1.0
- **Fix** (file and field you changed): in deployment.yaml, `image: hello:0.1.10` to `image: hello:0.1.1`

## Fault 3

- **Symptom** (what you saw first): `kubectl get pods` showing `CrashLoopBackOff` for the newest pod
- **Command** (that showed the cause): kubectl describe pod hello-6b57b75776-577wv
- **Evidence** (the exact output line):   Warning  Failed     100s (x7 over 7m4s)    kubelet            spec.containers{hello}: Error: failed to create containerd task: failed to create shim task: OCI runtime create failed: runc create failed: unable to start container process: container init was OOM-killed (memory limit too low?)
- **Cause**: memory limit too low
- **Fix** (file and field you changed): `memory: 2Mi` under `limits` changed to `memory: 62Mi` in deployment.yaml

## Fault 4

- **Symptom** (what you saw first): Newest pod in CrashLoopBackOff, this time with exit code 1 and Reason: Error
- **Command** (that showed the cause): kubectl logs hello-86cb4f567-mj7nr
- **Evidence** (the exact output line): 2026/11/19 09:30:34 listen tcp: lookup tcp/abc: unknown port 
- **Cause**: "abc" was set as a port, which doesn't exist
- **Fix** (file and field you changed): the value of the PORT environmental variable in deployment.yaml changed to "8080"

## Fault 5

- **Symptom** (what you saw first): Running but 0/1
- **Command** (that showed the cause): kubectl describe pod hello-6b7f5d9ff5-sfmnq
- **Evidence** (the exact output line): Warning  Unhealthy  2m14s (x19 over 3m14s)  kubelet            spec.containers{hello}: Readiness probe failed: Get "http://10.244.0.8:8081/healthz": dial tcp 10.244.0.8:8081: connect: connection refused
- **Cause**: The readiness probe asked for port 8081 but the server listens on 8080
- **Fix** (file and field you changed): in deployment.yaml, `port: 8081` changed to `port: 8080` under readinessProbe

## Fault 6

- **Symptom** (what you saw first): Running but 0/1
- **Command** (that showed the cause): kubectl describe pod hello-6b7f5d9ff5-sfmnq
- **Evidence** (the exact output line): Warning  Unhealthy  27s (x6 over 77s)   kubelet            spec.containers{hello}: Liveness probe failed: Get "http://10.244.0.8:8008/healthz": dial tcp 10.244.0.8:8008: connect: connection refused
- **Cause**: The liveness probe asked for port 8008 but the server listens on 8080
- **Fix** (file and field you changed): in deployment.yaml, `port: 8008` changed to `port: 8080` under livenessProbe

## Fault 7

- **Symptom** (what you saw first): The Service existed but EndpointSlice showed `ENDPOINTS <unset>` while both pods were 1/1
- **Command** (that showed the cause): kubectl describe svc hello and kubectl get pods --show-labels
- **Evidence** (the exact output line):
	Selector:                 app=hello,tier=web
	NAME                     READY   STATUS    RESTARTS   AGE   LABELS
	hello-84b887f966-fprj7   1/1     Running   0          11m   app=hello,pod-template-hash=84b887f966
- **Cause**: There was a `tier=web` label that pods don't have
- **Fix** (file and field you changed): removed `tier: web` from service.yaml

## Fault 8

- **Symptom** (what you saw first): The Service had two endpoints but requests through it didn't get a greeting
- **Command** (that showed the cause): kubectl get endpointslices
- **Evidence** (the exact output line): hello-ft9cv   IPv4          80      10.244.0.9,10.244.0.10   110m
- **Cause**: targetPort was 80, but the srever listens on 8080
- **Fix** (file and field you changed): changed `targetPort` in service.yaml to 8080
