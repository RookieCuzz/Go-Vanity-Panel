pipeline {
  agent {label 'dockeragent'}
  tools {
    go 'Go 1.23.6'              // 需提前在 Jenkins 全局工具里配置好 Go 安装
  }

  environment {
    GO111MODULE = 'on'        // 开启 Modules 模式
    CGO_ENABLED = '0'
    APP_NAME = 'vanityurl'
    REGISTRY = 'crpi-vqe38j3xeblrq0n4.cn-hangzhou.personal.cr.aliyuncs.com/go-mctown'
  }

  stages {
    stage('Checkout') {
      steps {
        checkout scm
      }
    }

    stage('Build') {
      steps {
        sh 'go mod tidy'
        sh 'go build -o ${env.APP_NAME}'
        sh 'ls -l' 
      }
    }


    stage('Package') {
      steps {
        sh 'tar czf ${env.APP_NAME}.tar.gz ${env.APP_NAME}'
      }
    }
    stage('Docker Build & Push') {
      steps {
        withCredentials([usernamePassword(
          credentialsId: 'aliyun-docker-login',
          usernameVariable: 'DOCKER_USERNAME',
          passwordVariable: 'DOCKER_PASSWORD'
        )]) {
          sh """
            echo "\$DOCKER_PASSWORD" | docker login --username \$DOCKER_USERNAME --password-stdin ${env.REGISTRY.split('/')[0]}
          """
        }

        script {
          def imageTag = "${env.REGISTRY}/${env.APP_NAME}:${env.BUILD_NUMBER}"
          def latestTag = "${env.REGISTRY}/${env.APP_NAME}::latest"

          sh """
            ls -l
            docker build -t ${imageTag} --network=host .
            docker tag ${imageTag} ${latestTag}
            docker push ${imageTag}
            docker push ${latestTag}
          """
        }
      }
    }

stage('Deploy All Compose Projects') {
      parallel {
        stage('Deploy compose1') {
  agent {label 'dockeragent'}
          steps {

              checkout scm
                              sh """
                pwd
                ls -l
                """
            dir('deploy/compose') {
              script {
                withCredentials([usernamePassword(
                  credentialsId: 'aliyun-docker-login',
                  usernameVariable: 'DOCKER_USERNAME',
                  passwordVariable: 'DOCKER_PASSWORD'
                )]) {
                  sh """
                    echo "$DOCKER_PASSWORD" | docker login --username "$DOCKER_USERNAME" --password-stdin ${env.REGISTRY.split('/')[0]}
                  """
                }
                sh """
                pwd
                ls -l
                docker compose -f docker-compose.yml down || true
                docker compose -f docker-compose.yml pull
                docker compose -f docker-compose.yml up -d --remove-orphans
                """
              }
            }
          }
        }

      }
    }


  }

  post {
    always {
      cleanWs()
    }
    success {
      echo "✅ 构建成功！"
    }
    failure {
      echo "🔥 构建失败，请检查日志。"
    }
  }
}
