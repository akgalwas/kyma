#!/usr/bin/env groovy
import groovy.json.JsonSlurperClassic
/*

Monorepo releaser: This Jenkinsfile runs the Jenkinsfiles of all subprojects based on the changes made and triggers kyma integration.
    - checks for changes since last successful build on master and compares to master if on a PR.
    - for every changed project, triggers related job async as configured in the seedjob.
    - for every changed additional project, triggers the kyma integration job.
    - passes info of:
        - revision
        - branch
        - current app version
        - all component versions

*/
def label = "kyma-${UUID.randomUUID().toString()}"
def registry = 'eu.gcr.io/kyma-project'
def acsImageName = 'acs-installer:0.0.4'
def changelogGeneratorPath = "tools/changelog-generator"

semVerRegex = /^([0-9]+\.[0-9]+\.[0-9]+)$/ // semVer format: 1.2.3
releaseBranchRegex = /^release\-([0-9]+\.[0-9]+)$/ // release branch format: release-1.5
isRelease = params.RELEASE_VERSION ==~ semVerRegex

commitID = ''
appVersion = ''
dockerPushRoot = ''

/*
    Projects that will be released.

    IMPORTANT NOTE: Projects trigger jobs and therefore are expected to have a job defined with the same name.
*/
projects = [
    "docs",
    "components/api-controller",
    "components/apiserver-proxy",
    "components/binding-usage-controller",
    "components/configurations-generator",
    "components/environments",
    "components/istio-webhook",
    "components/istio-kyma-patch",
    "components/helm-broker",
    "components/remote-environment-broker",
    "components/remote-environment-controller",
    "components/metadata-service",
    "components/gateway",
    "components/installer",
    "components/connector-service",
    "components/ui-api-layer",
    "components/event-bus",
    "components/event-service",
    "tools/alpine-net",
    "tools/watch-pods",
    "tools/stability-checker",
    "tools/etcd-backup",
    "tools/etcd-tls-setup",
    "tests/test-logging-monitoring",
    "tests/logging",
    "tests/acceptance",
    "tests/ui-api-layer-acceptance-tests",
    "tests/gateway-tests",
    "tests/test-environments",
    "tests/kubeless-test-client",
    "tests/api-controller-acceptance-tests",
    "tests/connector-service-tests",
    "tests/metadata-service-tests",
    "tests/remote-environment-controller-tests",
    "tests/event-bus"
]

/*
    project jobs to run are stored here to be sent into the parallel block outside the node executor.
*/
jobs = [:]

try {
    podTemplate(label: label) {
        node(label) {
            timestamps {
                ansiColor('xterm') {
                    stage("setup") {
                        checkout scm

                        // validate parameters
                        if (!isRelease && !params.RELEASE_VERSION.isEmpty()) {
                            error("Release version ${params.RELEASE_VERSION} does not follow semantic versioning.")
                        }
                        if (!params.RELEASE_BRANCH ==~ releaseBranchRegex) {
                            error("Release branch ${params.RELEASE_BRANCH} is not a valid branch. Provide a branch such as 'release-1.5'")
                        }
                    
                        commitID = sh (script: "git rev-parse origin/${params.RELEASE_BRANCH}", returnStdout: true).trim()
                        configureBuilds()
                    }

                    stage('collect projects') {
                        for (int i=0; i < projects.size(); i++) {
                            def index = i
                            jobs["${projects[index]}"] = { ->
                                    build job: "kyma/"+projects[index]+"-release",
                                            wait: true,
                                            parameters: [
                                                string(name:'GIT_REVISION', value: "$commitID"),
                                                string(name:'GIT_BRANCH', value: "${params.RELEASE_BRANCH}"),
                                                string(name:'APP_VERSION', value: "$appVersion"),
                                                string(name:'PUSH_DIR', value: "$dockerPushRoot"),
                                                booleanParam(name:'FULL_BUILD', value: true)
                                            ]
                            }
                        }
                    }
                }
            }
        }
    }

    // build components
    stage('build projects') {
        parallel jobs
    }

    // test the release
    stage('launch Kyma integration') {
        build job: 'kyma/integration-release',
            wait: true,
            parameters: [
                string(name:'GIT_REVISION', value: "$commitID"),
                string(name:'GIT_BRANCH', value: "${params.RELEASE_BRANCH}"),
                string(name:'APP_VERSION', value: "$appVersion")
            ]
    }

    // publish release artifacts
    podTemplate(label: label) {
        node(label) {
            timestamps {
                ansiColor('xterm') {
                    stage("setup") {
                        checkout scm
                    }

                    stage("Publish ${isRelease ? 'Release' : 'Prerelease'} ${appVersion}") {
                        def zip = "${appVersion}.tar.gz"
                        
                        // create release zip                        
                        sh "tar -czf ${zip} ./installation ./resources"

                        // create release on github
                        withCredentials(
                                [string(credentialsId: 'public-github-token', variable: 'token'),
                                sshUserPrivateKey(credentialsId: "bitbucket-rw", keyFileVariable: 'sshfile')
                            ]) {
                            
                            // Build changelog generator
                            dir(changelogGeneratorPath) {
                                sh "docker build -t changelog-generator ."
                            }   
                            
                            // Generate release changelog
                            changelogGenerator('/app/generate-release-changelog.sh --configure-git', ["LATEST_VERSION=${appVersion}", "GITHUB_AUTH=${token}", "SSH_FILE=${sshfile}"])

                            // Generate CHANGELOG.md
                            changelogGenerator('/app/generate-full-changelog.sh --configure-git', ["LATEST_VERSION=${appVersion}", "GITHUB_AUTH=${token}", "SSH_FILE=${sshfile}"])
                            sh "BRANCH=${params.RELEASE_BRANCH} LATEST_VERSION=${appVersion} SSH_FILE=${sshfile} APP_PATH=./tools/changelog-generator/app ./tools/changelog-generator/app/push-full-changelog.sh --configure-git"
                            commitID = sh (script: "git rev-parse HEAD", returnStdout: true).trim()

                            def releaseChangelog = readFile "./.changelog/release-changelog.md"
                            def body = releaseChangelog.replaceAll("(\\r|\\n|\\r\\n)+", "\\\\n")
                            def data = "'{\"tag_name\": \"${appVersion}\",\"target_commitish\": \"${commitID}\",\"name\": \"${appVersion}\",\"body\": \"${body}\",\"draft\": false,\"prerelease\": ${isRelease ? 'false' : 'true'}}'"
                            echo "Creating a new release using GitHub API..."
                            def json = sh (script: "curl --data ${data} -H \"Authorization: token $token\" https://api.github.com/repos/kyma-project/kyma/releases", returnStdout: true)
                            echo "Response: ${json}"
                            def releaseID = getGithubReleaseID(json)
                            // upload zip file
                            sh "curl --data-binary @$zip -H \"Authorization: token $token\" -H \"Content-Type: application/zip\" https://uploads.github.com/repos/kyma-project/kyma/releases/${releaseID}/assets?name=${zip}"                          
                        }
                    }
                }
            }
        }
    }
} catch (ex) {
    echo "Got exception: ${ex}"
    currentBuild.result = "FAILURE"
    def body = "${currentBuild.currentResult} ${env.JOB_NAME}${env.BUILD_DISPLAY_NAME}: on branch: ${env.BRANCH_NAME}. See details: ${env.BUILD_URL}"
    emailext body: body, recipientProviders: [[$class: 'DevelopersRecipientProvider'], [$class: 'CulpritsRecipientProvider'], [$class: 'RequesterRecipientProvider']], subject: "${currentBuild.currentResult}: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'"
}

/* -------- Helper Functions -------- */

/** Configure the parameters for the components to build:
 * - release candidate: push root: "rc/" / image tag: short commit ID
 * - release: push root: "" / image tag: semantic version
 */
def configureBuilds() {
    if(isRelease) {
        echo ("Building Release ${params.RELEASE_VERSION}")
        dockerPushRoot = ""
        appVersion = params.RELEASE_VERSION
    } else {
        echo ("Building Release Candidate for ${params.RELEASE_BRANCH}")
        dockerPushRoot = "rc/"
        appVersion = "${(params.RELEASE_BRANCH =~ /([0-9]+\.[0-9]+)$/)[0][1]}-rc" // release branch number + '-rc' suffix (e.g. 1.0-rc)
    }   
}

/**
 * Obtain the github release ID from its JSON data.
 * More info: https://developer.github.com/v3/repos/releases 
 */
@NonCPS
def getGithubReleaseID(releaseJson) {
    def slurper = new JsonSlurperClassic()
    return slurper.parseText(releaseJson).id
}

def changelogGenerator(command, envs = []) {
    def repositoryName = 'kyma'
    def image = 'changelog-generator'
    def envText = ''
    for (it in envs) {
        envText = "$envText --env $it"
    }
    workDir = pwd()

    def dockerRegistry = env.DOCKER_REGISTRY
    sh "docker run --rm -v $workDir:/$repositoryName -w /$repositoryName $envText $image sh $command"
}