# Mattermost AI Plugin

[![standard-readme compliant](https://img.shields.io/badge/readme%20style-standard-brightgreen.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

![Screenshot](/img/mention_bot.png)

## Table of Contents

- [Mattermost AI Plugin](#mattermost-ai-plugin)
  - [Table of Contents](#table-of-contents)
  - [Background](#background)
  - [Install](#install)
  - [Usage](#usage)
    - [Conversation](#conversation)
    - [Thread Summarization](#thread-summarization)
    - [Answer questions about Threads](#answer-questions-about-threads)
    - [Chat anywhere](#chat-anywhere)
    - [React for me](#react-for-me)
    - [RLHF Feedback Collection](#rlhf-feedback-collection)
  - [Related Efforts](#related-efforts)
  - [Contributing](#contributing)
  - [License](#license)

## Background

**🚀 Check out [our AI developer website](https://mattermost.github.io/mattermost-ai-site/) and join the ["AI Exchange" channel](https://community.mattermost.com/core/channels/ai-exchange) where Mattermost's open source community is sharing AI news and innovation in real time!**

The LLM extensions plugin adds functionality around the use and development of LLMs like GPT-3.5 / 4 and hugging face models within Mattermost. 

Currently at the experimental phase of development. Contributions and suggestions welcome! 

## Install

1. Clone and enter this repository:
  * `git clone https://github.com/mattermost/mattermost-plugin-ai && cd mattermost-plugin-ai`
2. Start the services: `docker compose up -d`
3. Configure the Mattermost server from the init script: `bash ./init.sh`
4. Access Mattermost:
  * Open Mattermost at `http://localhost:8065`
  * Select **View in Browser**
  * Log in with the generated `root` credentials
5. Install the `mattermost-plugin-ai` on Mattermost:
  * Go the releases page and download the latest release.
  * In the top left Mattermost menu, click **System Console** ➡️ **Plugin Management** and [upload the plugin to install it]((https://docs.mattermost.com/administration/plugins.html#plugin-uploads))
  * Enable the plugin and configure plugin settings as desired.

Lots of unfinished work in the system console settings. For now all you need to do is input and OpenAI API Key and configure allowed teams/users as desired. More options and the ability to use local LLMs is on the roadmap.

## Usage

### Conversation

Chat with an LLM right inside the Mattermost interface. Answer are streamed so you don't have to wait:

https://github.com/mattermost/mattermost-plugin-ai/assets/3191642/f375f1a2-61bf-4ae1-839b-07e44461809b

### Thread Summarization
Use the post menu or the `/summarize` command to get a summary of the thread in a DM:

![Summarizing Thread](/img/summarize_thread.png)

### Answer questions about Threads
Respond to the bot post to ask follow up questions:

https://github.com/mattermost/mattermost-plugin-ai/assets/3191642/6fed05e2-ee68-40db-9ee4-870c61ccf5dd

### Chat anywhere
Just mention @llmbot anywhere in Mattermost to ask it to respond. It will be given the context of the thread you are participating in:

![Bot Chat](/img/mention_bot.png)

### React for me
Just for fun! Use the post menu to ask the bot to react to the post. It will try to pick an appropriate reaction.

https://github.com/mattermost/mattermost-plugin-ai/assets/3191642/5282b066-86b5-478d-ae10-57c3cb3ba038

### RLHF Feedback Collection
Bot posts will have 👍 👎 icons for collecting feedback. The idea would be to use this as input for RLHF fine tuning.

## Related Efforts

Explore Mattermost's AI initiatives:

* https://mattermost.github.io/mattermost-ai-site/
* https://community.mattermost.com/core/channels/ai-exchange
* https://forum.mattermost.com/c/ai-frameworks/40
* https://mattermost.github.io/mattermost-ai-framework/
* https://docs.mattermost.com/about/mattermost-customizable-chatgpt-bot-framework.html
* https://mattermost.com/add-chatgpt-to-mattermost/
* https://github.com/Brightscout/mattermost-plugin-openai
* https://github.com/yGuy/chatgpt-mattermost-bot

## Contributing

Visit [our AI developer website](https://mattermost.github.io/mattermost-ai-site/) and check out Mattermost's [contributor guide](https://developers.mattermost.com/contribute/) to learn about contributing to our open source projects like this one.

## License

This repository is licensed under [Apache-2](./LICENSE).
