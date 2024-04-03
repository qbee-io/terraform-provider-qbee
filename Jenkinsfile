@Library('jenkins-shared-lib@7.0.0') _

pipeline {
    environment {
        ECR_REPO = "/booq/booq-jenkins-terraform-qbee"
    }

    agent {
        kubernetes {
            yaml '''
                apiVersion: v1
                kind: Pod
                metadata:
                labels:
                    jenkins/jenkins-jenkins-agent: true
                    jenkins/label: terraform-provider-qbee
                    app.kubernetes.io/name: terraform-provider-qbee
                spec:
                    containers:
                    - name: golang
                      image: golang:1.21-bookworm
                      command:
                      - cat
                      tty: true
                    volumes:
                    - name: docker-socket
                      hostPath:
                        path: /var/run/docker.sock
            '''
            nodeSelector 'instance-type=spot'
        }
    }

    stages {
        stage('Prepare') {
            steps {
                container('golang') {
                    script {
                        sh "git config --global --add safe.directory '*'"
                        env.PROVIDER_VERSION = sh(script: "git describe --abbrev=0 --tags", returnStdout: true).trim()
                    }
                }
            }
        }

        stage('Test') {
            steps {
                container('golang') {
                    sh 'go test ./...'
                }
            }
        }

        stage ('Release Build') {
            when {
                buildingTag()
            }

            steps {
                container('golang') {
                    sh 'curl -sfL https://goreleaser.com/static/run | bash -s -- release --skip-sign --clean'
                }
            }
        }

        stage('Jenkins Agent Container - Images') {
            parallel {
                stage('amd64') {
                    steps {
                        container('kaniko') {
                            buildAndPublishOCIImageWithKaniko("${ECR_REPO}", "amd64", " --build-arg TARGETARCH=amd64 --build-arg PROVIDER_VERSION=${env.PROVIDER_VERSION} --dockerfile=containers/JenkinsAgentDockerfile")
                        }
                    }
                    agent {
                        kubernetes {
                            customWorkspace 'build'
                            yaml '''
                        spec:
                            tolerations:
                            - key: "kaniko"
                              operator: "Equal"
                              value: "true"
                              effect: "NoSchedule"
                            nodeSelector:
                                kubernetes.io/arch: amd64
                            containers:
                            - name: kaniko
                              image: 386815924651.dkr.ecr.eu-west-1.amazonaws.com/kaniko-project/executor:v1.16.0-debug
                              imagePullPolicy: IfNotPresent
                              resources:
                                requests:
                                  memory: "8192Mi"
                              command:
                              - sleep
                              args:
                              - 99d
                    '''
                            nodeSelector 'instance-type=spot'
                        }
                    }
                }

                stage('arm64') {
                    steps {
                        container('kaniko') {
                            buildAndPublishOCIImageWithKaniko("${ECR_REPO}", "arm64", " --build-arg TARGETARCH=arm64 --build-arg PROVIDER_VERSION=${env.PROVIDER_VERSION} --dockerfile=containers/JenkinsAgentDockerfile")
                        }
                    }
                    agent {
                        kubernetes {
                            customWorkspace 'build'
                            yaml '''
                        spec:
                            tolerations:
                            - key: "kaniko"
                              operator: "Equal"
                              value: "true"
                              effect: "NoSchedule"
                            nodeSelector:
                                kubernetes.io/arch: arm64
                            containers:
                            - name: kaniko
                              image: 386815924651.dkr.ecr.eu-west-1.amazonaws.com/kaniko-project/executor:v1.16.0-debug
                              imagePullPolicy: IfNotPresent
                              resources:
                                requests:
                                  memory: "8192Mi"
                              command:
                              - sleep
                              args:
                              - 99d
                    '''
                            nodeSelector 'instance-type=spot'
                        }
                    }
                }
            }
        }

        stage('Jenkins Agent Container - OCI Manifest') {
            steps {
                container('manifest-tool') {
                    buildAndPublishOCIImageManifestWithManifestTool("${ECR_REPO}", "linux/arm64,linux/amd64 ")
                }
            }
            agent {
                kubernetes {
                    customWorkspace 'build'
                    yaml '''
                        spec:
                            nodeSelector:
                                kubernetes.io/arch: amd64
                            containers:
                            - name: manifest-tool
                              image: 386815924651.dkr.ecr.eu-west-1.amazonaws.com/manifest-tool:1.0.1-amd64
                              command:
                              - sleep
                              args:
                              - 99d
                    '''
                    nodeSelector 'instance-type=spot'
                }
            }
        }
    }
}
