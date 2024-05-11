<p align="center">
  <img src="https://raw.githubusercontent.com/Johnnycyan/Twitch-APIs/main/OneMoreDayIcon.svg" width="100" alt="project-logo">
</p>
<p align="center">
    <h1 align="center">AI-TWITCH-TTS</h1>
</p>
<p align="center">
    <em>Empower your streams with dynamic voice interactions.</em>
</p>
<p align="center">
	<img src="https://img.shields.io/github/last-commit/Johnnycyan/AI-Twitch-TTS?style=default&logo=git&logoColor=white&color=0080ff" alt="last-commit">
	<img src="https://img.shields.io/github/languages/top/Johnnycyan/AI-Twitch-TTS?style=default&color=0080ff" alt="repo-top-language">
	<img src="https://img.shields.io/github/languages/count/Johnnycyan/AI-Twitch-TTS?style=default&color=0080ff" alt="repo-language-count">
<p>
<p align="center">
	<!-- default option, no dependency badges. -->
</p>

<br><!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary><br>

- [ Overview](#-overview)
- [ Modules](#-modules)
- [ Getting Started](#-getting-started)
  - [ Installation](#-installation)
  - [ Usage](#-usage)
</details>
<hr>

##  Overview<a name="-overview"></a>

AI-Twitch-TTS is a real-time Twitch Text-to-Speech application built for interactive streaming experiences. The project orchestrates WebSocket connections for audio streaming, processes chat requests, and interfaces with external APIs for voice synthesis. It offers customizable voice options, real-time chat handling, and automated websocket reconnections, enhancing viewer engagement on Twitch streams. The projects modular design ensures a seamless integration of dependencies, automated testing, and CI/CD workflows for efficient development and deployment processes.

---

##  Modules<a name="-modules"></a>

| File                                                                             | Summary                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| ---                                                                              | ---                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| [index.html](https://github.com/Johnnycyan/AI-Twitch-TTS/blob/master/index.html) | Establish WebSocket connection for real-time audio streaming and processing. Decode audio data, apply gain, and compress dynamics before playback. Automatically reconnect on WebSocket closure and ensure proper closure on page exit.                                                                                                                                                                                                                                                                                                                        |                                                                                                                                                                                                                                                                                                                                             |
| [main.go](https://github.com/Johnnycyan/AI-Twitch-TTS/blob/master/main.go)       | This `main.go` file in the repository orchestrates a real-time Twitch Text-to-Speech (TTS) application. It handles client-server communication using WebSockets, processes chat requests, and interfaces with external APIs for voice synthesis. The code manages client connections, tracks request times, provides voice options, and sets default configurations for TTS alerts. It integrates modules for chat handling, random data generation, and environment variables loading, contributing to a responsive and interactive TTS experience on Twitch. |                                                                                                                                                                                                                                                                                                                                                                                            |

---

##  Getting Started<a name="-getting-started"></a>

**System Requirements:**

* **Internet**

###  Installation<a name="-installation"></a>

<h4>From <code>releases</code></h4>

> 1. Download latest release:
>     1. [Latest Release](https://github.com/Johnnycyan/AI-Twitch-TTS/releases)
>
> 2. Create a .env file in the same directory
>
> 3. Fill out required Environmental Variables explained below and in the .env.example

Variable      |     Description
------------- | -------------
ELEVENLABS_KEY  | Elevenlabs API key
SERVER_URL | URL of where the server will be hosted (no protocol) Ex: example.com
TTS_KEY | Secret key used to authenticate TTS generation
PALLY_KEYS | Json string list of name/key pairs for pally (optional)
VOICES | Json string list of name/id pairs for Elevenlabs voices

###  Usage<a name="-usage"></a>

<h4>From <code>releases</code></h4>

> ⚠️ Might not work without an SSL connection. Has not been tested.
> 1. Run AI-Twitch-TTS using the command below:
> ```console
> $ ./AI-Twitch-TTS
> ```
> or
> ```console
> $ AI-Twitch-TTS.exe
> ```
> 2. Add this to your OBS as a browser source
> ```
> http(s)://$SERVER_URL/?channel=<username>
> ```
> 3. Generate TTS by accessing this URL either through a browser or a Twitch chat bot (voice is optional and won't work unless building yourself anyways):
> ```
> http(s)://$SERVER_URL/tts?channel=<username>&key=$TTS_KEY&voice=<voicename>&text=<text to generate>
> ```

---

[**Return**](#-overview)

---
