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
	<img src="https://img.shields.io/github/license/Johnnycyan/AI-Twitch-TTS?logo=opensourceinitiative&logoColor=white&color=0080ff" alt="license">
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
- [ Example](#-example)
- [ Features](#-features)
- [ Modules](#-modules)
- [ Getting Started](#-getting-started)
  - [ Installation](#-installation)
  - [ Usage](#-usage)
  - [ Advanced Usage](#-advanced-usage)
- [ License](#-license)
</details>
<hr>

<a name="-overview"></a>
##  Overview

AI-Twitch-TTS is a real-time Twitch Text-to-Speech application built for interactive streaming experiences. The project orchestrates WebSocket connections for audio streaming, processes chat requests, and interfaces with external APIs for voice synthesis. It offers customizable voice options, real-time chat handling, and automated websocket reconnections, enhancing viewer engagement on Twitch streams. The projects modular design ensures a seamless integration of dependencies, automated testing, and CI/CD workflows for efficient development and deployment processes.

---

##  Example Usage from [Samifying](https://www.twitch.tv/samifying)<a name="-example"></a>

https://github.com/Johnnycyan/AI-Twitch-TTS/assets/24556317/3996ecab-cb1e-4e46-9964-2773146901d8


##  Features<a name="-features"></a>

|    |   Feature         | Description |
|----|-------------------|---------------------------------------------------------------|
| ⚙️  | **Architecture**  | Server-side application using WebSockets for real-time audio streaming, with client-side support for Twitch Text-to-Speech functionality. Maintains web server to handle requests and WebSocket connections effectively. |
| 🔩 | **Code Quality**  | Well-structured codebase with clear separation of concerns, detailed inline comments, consistent naming conventions, and adherence to best practices. Follows the principles of clean code and maintainable architecture. |
| 📄 | **Documentation** | Adequate documentation with detailed explanations for modules like logging, environment setup, WebSocket handling, and HTTP endpoints. |
| 🔌 | **Integrations**  | Relies on external libraries like godotenv, go-randomdata, WebSocket for Go, and others to enhance functionality like environment variable loading, random data generation, WebSocket communication, and real-time audio streaming. |
| 🧩 | **Modularity**    | Codebase exhibits modularity through separate modules for logging, WebSocket handling, text-to-speech requests, alerts retrieval, and Pally WebSocket connections. Modules are designed for reusability and maintainability. |

---

##  Modules<a name="-modules"></a>

| File                                                                                     | Summary                                                                                                                                                                                                                                                                                        |
| ---                                                                                      | ---                                                                                                                                                                                                                                                                                            |
| [alerts.go](https://github.com/Johnnycyan/AI-Twitch-TTS/blob/master/alerts.go)           | Retrieves and delivers random alert sounds based on channel. Checks for available sound files in the designated folder and selects one at random. Handles errors in case of missing or inaccessible files.                                                                                     |
| [environment.go](https://github.com/Johnnycyan/AI-Twitch-TTS/blob/master/environment.go) | Loads environment variables using godotenv, ensuring essential values are present. Sets up necessary configurations for the AI-Twitch-TTS application to function correctly, maintaining a robust system.                                                                                      |
| [logging.go](https://github.com/Johnnycyan/AI-Twitch-TTS/blob/master/logging.go)         | Enables logging customization based on user-defined levels to ensure relevant messages are displayed according to set verbosity preferences.                                                                                                                                                   |
| [main.go](https://github.com/Johnnycyan/AI-Twitch-TTS/blob/master/main.go)               | Defines HTTP handlers for text-to-speech and websockets, serving HTML template. Orchestrates server setup and logging, listening on specified port.                                                                                                                                            |
| [pally.go](https://github.com/Johnnycyan/AI-Twitch-TTS/blob/master/pally.go)             | Establishes and manages connections to [Pally](https://pally.gg) WebSocket, processing tip notifications for a Twitch extension. Handles message parsing and client interactions, ensuring timely and accurate message delivery. Maintains WebSocket communication and connection stability for real-time updates. |
| [tts.go](https://github.com/Johnnycyan/AI-Twitch-TTS/blob/master/tts.go)                 | Handles text-to-speech requests by generating audio data and sending it to connected clients based on specified channels and voices. Manages voice configurations, client connections, rate limits, and authorization keys for seamless TTS functionality.                                     |
| [websockets.go](https://github.com/Johnnycyan/AI-Twitch-TTS/blob/master/websockets.go)   | Implements WebSocket handling for client connections, managing periodic pings, and message processing. Dynamically assigns client names, tracks active clients, and logs events. Enhances real-time communication in the AI-Twitch-TTS repository architecture.                                |
| [index.html](https://github.com/Johnnycyan/AI-Twitch-TTS/blob/master/index.html)         | Implements real-time audio streaming via WebSocket in the AI-Twitch-TTS client. Handles WebSocket connection, audio processing, pings, and restarts. Auto-reconnects on close.                                                                                                                 |

---

##  Getting Started<a name="-getting-started"></a>

**System Requirements:**

* **Internet**

###  Installation<a name="-installation"></a>

<h4>From <code>releases</code></h4>

> 1. Download latest release:
>     1. [Latest Release](https://github.com/Johnnycyan/AI-Twitch-TTS/releases/latest)
>
> 2. Create ./alerts/channel folder with alert sound(s) in it for [Pally](https://pally.gg) (optional)
>
> 3. Create ./effects folder with effect sound(s) in it for effect tags
>
> 4. Create a .env file in the same directory
>
> 5. Fill out required Environmental Variables explained below and in the .env.example

Variable      |     Description
------------- | -------------
ELEVENLABS_KEY  | Elevenlabs API key
SERVER_URL | URL of where the server will be hosted (no protocol) Ex: example.com
TTS_KEY | Secret key used to authenticate TTS generation
VOICES | Json string list of name/id pairs for Elevenlabs voices
PALLY_KEYS | Json string list of name/key pairs for [Pally](https://pally.gg) (optional)
SENTRY_URL | URL for Sentry logging of the client (optional)

##  Usage<a name="-usage"></a>

<h4>From <code>releases</code></h4>

> ⚠️ Might not work without an SSL connection. Has not been tested.
> 1. Run AI-Twitch-TTS using the command below:
>     1. Logging mode is optional. Options: info, debug, fountain
> ```console
> $ ./AI-Twitch-TTS <port> <logging-mode>
> ```
> or
> ```console
> $ AI-Twitch-TTS.exe <port> <logging-mode>
> ```
> 2. Add this to your OBS as a browser source
> ```
> http(s)://$SERVER_URL/?channel=<username>
> ```
> 3. Generate TTS by accessing this URL either through a browser or a Twitch chat bot (voice is optional):
>     1. See [ Advanced Usage](#-advanced-usage) to see how to use multiple voices and effects in one message.
> ```
> http(s)://$SERVER_URL/tts?channel=<username>&key=$TTS_KEY&voice=<voicename>&text=<text to generate>
> ```

---

##  Advanced Usage<a name="-advanced-usage"></a>

[v-voicename] is a voice tag meaning any text written after it will be spoken with that voice.

[e-effectname] is an effect tag which will play an effect.

If you use a tag in a message you MUST use voice tags for all the text you want to say.

✅`[v-voice] this is text and then an effect [e-effect]`

❌`this text has no voice tag [e-effect]`

✅`[v-voice] this is text and then an effect [e-effect] [v-voice] and then more text`

❌`[v-voice] this is text and then an effect [e-effect] this text has no voice tag`

✅`[e-effect] [v-voice] this is text`

❌`[e-effect] this text has no voice tag`

---

##  License<a name="-license"></a>

This project is protected under the [MIT](https://choosealicense.com/licenses/mit/) License. For more details, refer to the [LICENSE](https://github.com/Johnnycyan/AI-Twitch-TTS/blob/master/LICENSE) file.

---

[**Return**](#-overview)
