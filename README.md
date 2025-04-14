# Poster-Setter

## Description

**Poster-Setter** is a tool designed to use Mediux images for Plex content. It provides a simple and intuitive web interface to browse and select image sets for your Plex library. While there are many tools available for this purpose, Poster-Setter stands out by offering a visual web interface to preview images before selection.

> **Note:** This tool is currently in beta. While it has been tested locally, please use it at your own risk.

## Features

-   **Mediux API Integration**: Seamlessly fetch and use Mediux images.
-   **Web GUI**: Preview and select image sets through an easy-to-use web interface.
-   **Automatic Updates**: Save sets to a local SQLite database for periodic updates.
-   **Local Image Storage**: Option to save downloaded images next to your Plex content.

## Installation

Poster-Setter is designed to run in Docker for easy setup and deployment.

1. Clone the repository:
    ```sh
    git clone https://github.com/xmoosex/poster-setter.git
    ```
2. Navigate to the project directory:
    ```sh
    cd poster-setter
    ```
3. Tweak the docker-compose file to match your settings
4. Run the application with:
    ```sh
    docker-compose up --build
    ```

Before running the application in Docker, you need to set up a config.yaml file. You can use the config.yaml.sample file as a template.

## Usage

1. Access the web interface by navigating to `http://localhost:8888` in your browser.
2. Browse all of your Plex content and choose what you want to search Mediux for.
3. Browse and preview Mediux image sets for that Plex item.
4. Select the set you want to use for your Plex content.
5. Choose what you want to download from that set (eg: Poster, Backdrop, Season Posters, Titlecards)
6. Save sets for automatic updates (optional)

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
