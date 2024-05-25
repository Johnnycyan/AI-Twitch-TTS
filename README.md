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

<a name="-example"></a>

##  Example Usage from [Samifying](https://www.twitch.tv/samifying)

https://github.com/Johnnycyan/AI-Twitch-TTS/assets/24556317/3996ecab-cb1e-4e46-9964-2773146901d8

---

##  Features<a name="-features"></a>


|    |   Feature         | Description |
|----|-------------------|---------------------------------------------------------------|
| ⚙️  | **Architecture**  | Server-side application using WebSockets for real-time audio streaming, with client-side support for Twitch Text-to-Speech functionality. Maintains web server to handle requests and WebSocket connections effectively. |
| 🔩 | **Code Quality**  | Well-structured codebase with clear separation of concerns, detailed inline comments, consistent naming conventions, and adherence to best practices. Follows the principles of clean code and maintainable architecture. |
| 📄 | **Documentation** | Adequate documentation with detailed explanations for modules like logging, environment setup, WebSocket handling, and HTTP endpoints. |
| 🔌 | **Integrations**  | Relies on external libraries like godotenv, go-randomdata, WebSocket for Go, and others to enhance functionality like environment variable loading, random data generation, WebSocket communication, and real-time audio streaming. |
| 🧩 | **Modularity**    | Codebase exhibits modularity through separate modules for logging, WebSocket handling, text-to-speech requests, alerts retrieval, and Pally WebSocket connections. Modules are designed for reusability and maintainability. |

---

<a name="-modules"></a>

##  Modules

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

<a name="-getting-started"></a>

##  Getting Started

**System Requirements:**

* **Internet**

<a name="-installation"></a>

###  Installation

<h4>From <code>releases</code></h4>

> 1. Download latest release:
>     1. [Latest Release](https://github.com/Johnnycyan/AI-Twitch-TTS/releases/latest)
>
> 2. Create `./alerts/<channel>` folder with alert sound(s) in it for [Pally](https://pally.gg) (optional)
>
> 3. Create `./effects` folder with effect sound(s) in it for effect tags
>
> 4. Create a `.env` file in the same directory
>
> 5. Fill out required Environmental Variables explained below and in the .env.example

---

<h4>From <code>docker</code></h4>

> 1. Create `./effects` folder with effect sound(s) in it for effect tags
>
> 2. Create `./alerts/<channel>` folder with alert sound(s) in it for [Pally](https://pally.gg) (optional)
>
> 3. Either create a `.env` file with the required Environmental Variables explained below and in the .env.example or just change them in the compose file.

`docker-compose.yml`
```
version: "3.8"
services:
  ai-twitch-tts:
    image: johnnycyan/ai-twitch-tts:main
    container_name: tts
    ports:
      - 6969:8080
    environment:
      - ELEVENLABS_KEY=${ELEVENLABS_KEY}
      - ELEVENLABS_PRICE=${ELEVENLABS_PRICE}
      - SERVER_URL=${SERVER_URL}
      - SENTRY_URL=${SENTRY_URL}
      - TTS_KEY=${TTS_KEY}
      - PALLY_KEYS=${PALLY_KEYS}
      - VOICES=${VOICES}
      - VOICE_MODELS=${VOICE_MODELS}
      - VOICE_STYLES=${VOICE_STYLES}
      - VOICE_MODIFIERS=${VOICE_MODIFIERS}
      - MONGO_HOST=mongodb
      - MONGO_PORT=27017
      - MONGO_USER=${MONGO_USER}
      - MONGO_PASS=${MONGO_PASS}
      - MONGO_DB=${MONGO_DB}
      - FFMPEG_ENABLED=true
    volumes:
      - ./effects:/app/effects
      - ./alerts:/app/alerts
    depends_on:
      - mongodb
  mongodb:
    image: mongo
    container_name: tts-mongo
    restart: always
    environment:
      - MONGO_INITDB_ROOT_USERNAME=${MONGO_USER}
      - MONGO_INITDB_ROOT_PASSWORD=${MONGO_PASS}
    volumes:
      - mongodb_data:/data/db
volumes:
  mongodb_data:
```

Variable         |  Description
-------------    | -------------
ELEVENLABS_KEY   | Elevenlabs API key
SERVER_URL       | URL of where the server will be hosted (no protocol) Ex: example.com
TTS_KEY          | Secret key used to authenticate TTS generation
VOICES           | Json string list of name/id pairs for Elevenlabs voices
VOICE_MODELS     | Json string list of name/model pairs for Elevenlabs voices (optional)
VOICE_STYLES     | Json string list of name/style pairs for Elevenlabs voices (optional)
VOICE_MODIFIERS  | Json string list of name/modifier pairs for Elevenlabs voices (optional)
PALLY_KEYS       | Json string list of name/key pairs for [Pally](https://pally.gg) (optional)
SENTRY_URL       | URL for Sentry logging of the client (optional)
MONGO_HOST       | URL for MongoDB Host (optional)
MONGO_PORT       | Port for MongoDB (optional)
MONGO_USER       | Username for MongoDB (optional)
MONGO_PASS       | Password for MongoDB (optional)
MONGO_DB         | Database name for MongoDB (optional)
ELEVENLABS_PRICE | Monthly Price of Elevenlabs Subscription (optional)
FFMPEG_ENABLED   | Bool for if you have ffmpeg installed for voice modifiers (optional)

<a name="-usage"></a>

##  Usage

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

<h4>From <code>docker</code></h4>

> ⚠️ Might not work without an SSL connection. Has not been tested.
> 1. Add this to your OBS as a browser source
> ```
> http(s)://$SERVER_URL/?channel=<username>
> ```
> 2. Generate TTS by accessing this URL either through a browser or a Twitch chat bot (voice is optional):
>     1. See [ Advanced Usage](#-advanced-usage) to see how to use multiple voices and effects in one message.
> ```
> http(s)://$SERVER_URL/tts?channel=<username>&key=$TTS_KEY&voice=<voicename>&text=<text to generate>
> ```

---

<a name="-advanced-usage"></a>

##  Advanced Usage

[v-voicename] is a voice tag meaning any text written after it will be spoken with that voice.

[e-effectname] is an effect tag which will play an effect.

(reverb) adds reverb to a TTS message.

If you use a tag in a message you MUST use voice tags for all the text you want to say.

✅`[v-voice] this is text and then an effect [e-effect]`

❌`this text has no voice tag [e-effect]`

✅`[v-voice] this is text and then an effect [e-effect] [v-voice] and then more text`

❌`[v-voice] this is text and then an effect [e-effect] this text has no voice tag`

✅`[e-effect] [v-voice] this is text`

❌`[e-effect] this text has no voice tag`

Example of reverb:

`(reverb) this is reverbed text`

`[v-voice] (reverb) this is reverbed text with a specific voice.`

---

<a name="-license"></a>

##  License

This project is protected under the [MIT](https://choosealicense.com/licenses/mit/) License. For more details, refer to the [LICENSE](https://github.com/Johnnycyan/AI-Twitch-TTS/blob/master/LICENSE) file.

---

[**Return**](#-overview)
