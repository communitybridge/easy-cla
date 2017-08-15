def buildStatus = 0

node {
  properties([disableConcurrentBuilds()])

  // Wipe the workspace so we are building completely clean
  sh "sudo rm -rf *"

  stage ("Checkout") {
    git pool: true, credentialsId: 'd78c94c4-9179-4765-9851-9907b5ef2cc4', url: "git@github.linuxfoundation.org:Engineering/project-management-console.git", branch: "${env.BRANCH_NAME}"
  }

  def project = "pmc"
  def gitCommit = sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
  def shortCommit = gitCommit.take(7)
  def gitAuthor = sh(returnStdout: true, script: 'git log -1 --pretty=format:"%an"').trim()
  def gitCommitMsg = sh(returnStdout: true, script: 'git log -1 --pretty=format:"%s"').trim()

  try {

    stage ("Launching CINCO Instance") {
      dir('keycloak') {
        git branch: 'master', credentialsId: 'd78c94c4-9179-4765-9851-9907b5ef2cc4', url: 'git@github.linuxfoundation.org:Engineering/keycloak.git'
        sh "lf init -d"
        sh "sleep 60"
      }
      dir('cinco') {
        git branch: 'develop', credentialsId: 'd78c94c4-9179-4765-9851-9907b5ef2cc4', url: 'git@github.linuxfoundation.org:Engineering/integration-platform.git'
        sh "lf init -d --dep-map=keycloak:../keycloak/"
      }
    }

    stage ("Launching PMC Instance") {
      sh "lf init -d --mode=ci --dep-map=cinco:cinco/"
    }

    stage ("Waiting for CINCO") {
      timeout(10) {
        sh 'lf run wait-for-cinco'
      }
    }

    def workspaceID = sh (script: "lf workspace", returnStdout: true).trim()

    stage ("NPM Installation") {
      sh "docker exec ${workspaceID} bash -c \"cd src && npm install\""
    }

    stage ("Ionic Installation") {
      sh "docker exec ${workspaceID} bash -c \"cd src && npm run build\""
    }

    stage("Destroying Instances") {
      sh "lf -i cinco/ rm -y"
      sh "lf -i keycloak/ rm -y"
      sh "lf rm -y"
    }

    if (env.BRANCH_NAME == 'develop') {
      build job: 'PMC - Sandbox', parameters: [string(name: 'SHA', value: "${shortCommit}")], wait: false
    } else if (env.BRANCH_NAME == 'master') {
      build job: 'PMC - Production', parameters: [string(name: 'SHA', value: "${shortCommit}")], wait: false
    }

  } catch(err) {
    buildStatus = 1

    // Making sure we always destroy the instance after each build
    stage("Destroying Instances") {
      sh "lf -i cinco/ rm -y"
      sh "lf -i keycloak/ rm -y"
      sh "lf rm -y"
    }

    throw err

  } finally {
    withCredentials([string(credentialsId: 'workflow-api-key', variable: 'API_KEY')]) {
      sh "curl -s -H \"x-api-key: $API_KEY\" https://workflow.eng.linuxfoundation.org/trigger/jenkins/build_notif -d '{ \
        \"build\": \"${env.BUILD_ID}\", \
        \"build_url\": \"${env.BUILD_URL}\", \
        \"gitAuthor\": \"${gitAuthor}\", \
        \"gitCommit\": \"${shortCommit}\", \
        \"gitCommitMsg\": \"${gitCommitMsg}\", \
        \"job_name\": \"${env.JOB_NAME}\", \
        \"job_url\": \"${env.JOB_URL}\", \
        \"status\": \"${buildStatus}\", \
        \"gitBranch\": \"${env.BRANCH_NAME}\", \
        \"job_base_name\": \"${env.JOB_BASE_NAME}\", \
        \"channel\": \"#lfplatform-pmc\"}'"
    }
  }
}

