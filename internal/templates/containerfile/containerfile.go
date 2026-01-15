package containerfile

import (
	"strings"
	"text/template"
)

const (
	templateHeader = `# =============================================================================
# Android SDK Build Container
# =============================================================================
# Minimal Ubuntu-based container with Android SDK, Java, and Gradle for
# building Android applications in CI/CD or local development environments.
#
# Usage:
#   podman build -t android-sdk .
#   podman run -v $(pwd):/opt/sources android-sdk gradle build
#   # Or use the wrapper: podman run -v $(pwd):/opt/sources android-sdk ./gradlew build
# =============================================================================

`

	templateVersions = `# -----------------------------------------------------------------------------
# Version Configuration
# Modify these ARGs to customize installed versions
# -----------------------------------------------------------------------------
ARG JDK_VERSION={{.Versions.Jdk}}
ARG ANDROID_API_LEVEL={{.Versions.AndroidApiLevel}}
ARG ANDROID_BUILD_TOOLS_VERSION={{.Versions.AndroidBuildTools}}
ARG GRADLE_VERSION={{.Versions.Gradle}}
ARG ANDROID_CMDLINE_TOOLS_VERSION={{.Versions.AndroidCmdlineTools}}

`

	templateBaseImage = `# -----------------------------------------------------------------------------
# Base Image
# -----------------------------------------------------------------------------
FROM {{.Image}}

# Re-declare ARGs after FROM to make them available in build
ARG JDK_VERSION
ARG ANDROID_API_LEVEL
ARG ANDROID_BUILD_TOOLS_VERSION
ARG GRADLE_VERSION
ARG ANDROID_CMDLINE_TOOLS_VERSION

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

`

	templateSystemDeps = `# -----------------------------------------------------------------------------
# Install System Dependencies
# -----------------------------------------------------------------------------
RUN apt update && \
    apt install -y \
    wget \
    unzip \
    openjdk-$JDK_VERSION-jdk \
    && apt clean \
    && rm -rf /var/lib/apt/lists/*

`

	templateJava = `# -----------------------------------------------------------------------------
# Java Environment Setup
# -----------------------------------------------------------------------------
ENV JAVA_HOME=/usr/lib/jvm/java-$JDK_VERSION-openjdk-amd64
ENV PATH=$JAVA_HOME/bin:$PATH

`

	templateAndroidSdk = `# -----------------------------------------------------------------------------
# Install Android SDK Command Line Tools
# -----------------------------------------------------------------------------
RUN mkdir -p /opt/android-sdk/cmdline-tools && \
    wget -q https://dl.google.com/android/repository/commandlinetools-linux-${ANDROID_CMDLINE_TOOLS_VERSION}_latest.zip \
      -O /tmp/cmdline-tools.zip && \
    unzip -q /tmp/cmdline-tools.zip -d /tmp && \
    mv /tmp/cmdline-tools /opt/android-sdk/cmdline-tools/latest && \
    rm /tmp/cmdline-tools.zip

ENV ANDROID_HOME=/opt/android-sdk
ENV ANDROID_SDK_ROOT=/opt/android-sdk
ENV PATH=$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools:$ANDROID_HOME/emulator:$PATH

`

	templateAndroidComponents = `# -----------------------------------------------------------------------------
# Install Android SDK Components
# -----------------------------------------------------------------------------
RUN yes | sdkmanager --licenses && \
    sdkmanager --update && \
    sdkmanager \
    "platform-tools" \
    "platforms;android-$ANDROID_API_LEVEL" \
    "build-tools;$ANDROID_BUILD_TOOLS_VERSION" \
    "extras;android;m2repository" \
    "extras;google;m2repository"

`

	templateGradle = `# -----------------------------------------------------------------------------
# Install Gradle
# -----------------------------------------------------------------------------
RUN wget -q https://services.gradle.org/distributions/gradle-$GRADLE_VERSION-bin.zip \
      -O /tmp/gradle.zip && \
    unzip -q /tmp/gradle.zip -d /opt && \
    mv /opt/gradle-$GRADLE_VERSION /opt/gradle && \
    rm /tmp/gradle.zip

ENV GRADLE_HOME=/opt/gradle
ENV PATH=$GRADLE_HOME/bin:$PATH

`

	templateWorkdir = `# -----------------------------------------------------------------------------
# Setup Working Directory
# -----------------------------------------------------------------------------
RUN mkdir -p /opt/sources
WORKDIR /opt/sources

`

	templateCleanup = `# -----------------------------------------------------------------------------
# Cleanup
# -----------------------------------------------------------------------------
RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

`

	templateCmd = `# -----------------------------------------------------------------------------
# Default Command
# -----------------------------------------------------------------------------
CMD ["/bin/bash"]
`
)

type ContainerOpts struct {
	Image    string
	Versions Versions
}

type Versions struct {
	Jdk                 string
	AndroidApiLevel     string
	AndroidBuildTools   string
	Gradle              string
	AndroidCmdlineTools string
}

func NewContainerOpts() ContainerOpts {
	return ContainerOpts{
		Image: "ubuntu:24.04",
		Versions: Versions{
			Jdk:                 "21",
			AndroidApiLevel:     "36",
			AndroidBuildTools:   "36.1.0",
			Gradle:              "9.2.1",
			AndroidCmdlineTools: "13114758",
		},
	}
}

func GenerateContainerfile() (string, error) {
	return GenerateContainerfileWithOpts(NewContainerOpts())
}

func GenerateContainerfileWithOpts(opts ContainerOpts) (string, error) {
	containerfileTemplate := templateHeader +
		templateVersions +
		templateBaseImage +
		templateSystemDeps +
		templateJava +
		templateAndroidSdk +
		templateAndroidComponents +
		templateGradle +
		templateWorkdir +
		templateCleanup +
		templateCmd

	tmpl, err := template.New("containerfile").Parse(containerfileTemplate)
	if err != nil {
		return "", err
	}

	w := &strings.Builder{}
	err = tmpl.Execute(w, opts)
	if err != nil {
		return "", err
	}
	return w.String(), nil
}
