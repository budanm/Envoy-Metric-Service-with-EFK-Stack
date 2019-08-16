# Envoy-Metric-Service-with-EFK-Stack
A simple example to demonstrate the usage of envoy.metrics_service StatsSink.

## Description
The detailed description of this setup can be found this article https://link.medium.com/q7KExpSgcZ


## Instructions to run the project ( docker-compose )

The project comes with a docker-compose file which can be used as it is

Step1: Build the project
```
docker-compose build
```

Step2: Bring up the envoy containers using docker-compose
```
docker-compose up  
```

## Setup
The particular setup can be briefly summarized in the diagram as shown below
![Envoy stats monitoring with EFK stack](setup_diagram.png?raw=true "Deployment diagram")
