# PetNeed.Me LINE Bot

[![GoDoc](https://godoc.org/github.com/kkdai/LineBotPetNeedMe?status.svg)](https://godoc.org/github.com/kkdai/LineBotPetNeedMe)
[![Go](https://github.com/kkdai/LineBotPetNeedMe/workflows/Go/badge.svg)](https://github.com/kkdai/LineBotPetNeedMe/actions/workflows/go.yml)

This is a LINE bot that provides information about animals available for adoption from Taipei's animal shelters.

## Features

*   **Find Pets:** Search for cats and dogs available for adoption.
*   **Pet Information:** Get details about each pet, including a photo.
*   **Easy to Use:** Simply add the bot on LINE and start searching.

## Tech Stack

*   **Golang:** The bot is written in Go.
*   **LINE Bot SDK:** Uses the [line-bot-sdk-go](https://github.com/line/line-bot-sdk-go) library.

## Getting Started

To run the bot locally, you'll need to have Go installed and set up a LINE developer account.

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/kkdai/LineBotPetNeedMe.git
    ```
2.  **Install dependencies:**
    ```bash
    go get
    ```
3.  **Set up your environment variables:**
    *   `ChannelSecret`: Your LINE bot's channel secret.
    *   `ChannelAccessToken`: Your LINE bot's channel access token.
    *   `PORT`: The port you want to run the bot on (e.g., 8080).
    *   `IMG_SRV`: The image server URL.
4.  **Run the bot:**
    ```bash
    go run main.go
    ```

## Deployment

You can deploy this bot to Heroku using the button below:

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## License

This project is licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for details.

## Acknowledgments

*   Thanks to the [g0v](http://g0v.tw/) community for their inspiration and support.
*   Data is sourced from the [Taipei Open Data API](http://data.taipei/opendata/datalist/datasetMeta/outboundDesc?id=6a3e862a-e1cb-4e44-b989-d35609559463&rid=f4a75ba9-7721-4363-884d-c3820b0b917c).
*   Inspired by the [petneed.me](https://github.com/jsleetw/petneed.me) project.