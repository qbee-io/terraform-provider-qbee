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
                    sh "git config --global --add safe.directory '*'"
                }
            }
        }

//        stage('Test') {
//            steps {
//                container('golang') {
//                    sh 'go test ./...'
//                }
//            }
//        }
//
//        stage ('Release') {
//            when {
//                buildingTag()
//            }
//
//            steps {
//                container('golang') {
//                    sh 'curl -sfL https://goreleaser.com/static/run | bash -s -- release --skip-sign --clean'
//                }
//            }
//        }

        stage('Jenkins Agent Container - Images') {
            parallel {
                stage('amd64') {
                    steps {
                        container('kaniko') {
                            buildAndPublishImage("${ECR_REPO}", "amd64")
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
                            buildAndPublishImage("${ECR_REPO}", "arm64")
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

def buildAndPublishImage(ecrRepo, architecture) {
    def IMAGE_VERSION
    def IMAGE_TAG
    def AWS_ACCOUNT_ID="386815924651"
    def AWS_REGION="eu-west-1"
    def ECR_URL="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
    def SNAPSHOTS_ECR_URL="${ECR_URL}/snapshots${ecrRepo}"
    def RELEASES_ECR_URL="${ECR_URL}/releases${ecrRepo}"
    def BRANCH_NAME_UNDERSCORE=sh(script: "echo ${env.BRANCH_NAME} | tr '/' '_'", returnStdout: true).trim()
    def GIT_COMMIT_SHORT=sh(script: "echo ${GIT_COMMIT} | cut -c1-8", returnStdout: true).trim()

    echo "ARCHITECTURE: ${architecture}"
    echo "SNAPSHOTS_ECR_URL: ${SNAPSHOTS_ECR_URL}"
    echo "RELEASES_ECR_URL: ${RELEASES_ECR_URL}"
    echo "BRANCH_NAME_UNDERSCORE: ${BRANCH_NAME_UNDERSCORE}"
    echo "GIT_COMMIT_SHORT: ${GIT_COMMIT_SHORT}"

    def PROVIDER_VERSION=sh(script: "git describe --abbrev=0 --tags", returnStdout: true).trim()

    if(env.TAG_NAME == null) {
        // SNAPSHOT
        IMAGE_VERSION=sh(script: "echo ${BRANCH_NAME_UNDERSCORE}_${GIT_COMMIT_SHORT}", returnStdout: true).trim()
        echo "IMAGE_VERSION: ${IMAGE_VERSION}"
        echo "PROVIDER_VERSION: ${PROVIDER_VERSION}"

        def ARGS = " --build-arg ARCH=${architecture} --build-arg BUILD_RELEASE=false --build-arg VERSION=${PROVIDER_VERSION}"

        sh """
            /kaniko/executor --dockerfile=containers/JenkinsAgentDockerfile \
                             --context=`pwd` \
                             --destination=${SNAPSHOTS_ECR_URL}:${IMAGE_VERSION}-${architecture} \
                             --destination=${SNAPSHOTS_ECR_URL}:${BRANCH_NAME_UNDERSCORE}_LATEST-${architecture} \
                             ${ARGS}
        """
    } else {
        // RELEASE
        IMAGE_VERSION="${env.TAG_NAME}"
        echo "IMAGE_VERSION: ${IMAGE_VERSION}"

        def ARGS = " --build-arg ARCH=${architecture} --build-arg BUILD_RELEASE=true --build-arg VERSION=${PROVIDER_VERSION}"

        sh """
            /kaniko/executor --dockerfile=containers/JenkinsAgentDockerfile \
                             --context=`pwd` \
                             --destination=${RELEASES_ECR_URL}:${IMAGE_VERSION}-${architecture} \
                             ${ARGS}
        """
    }
}
