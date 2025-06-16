---
layout: default
title: "Configuration"
nav_order: 2
description: "Configuration instructions for AURA."
permalink: /config
---

# Configuration

AURA uses a `config.yml` file for configuration. This file is essential for setting up the application according to your preferences and environment. Below are the steps to create and modify the `config.yml` file:

1. **Create the `config.yml` File**:

    - You can create a new file named `config.yml` in the root directory of your AURA installation.
    - Alternatively, you can use the sample configuration file provided in the repository as a starting point. You can find it [here](https://raw.githubusercontent.com/mediux-team/aura/master/config.yml.sample).

2. **Edit the `config.yml` File**:
    - Open the `config.yml` file in a text editor of your choice.
    - Modify the configuration settings according to your needs.
3. **Place the `config.yml` File**:
    - Place your configuration file in the `/config` directory on your Docker container.

# Configuration Options

## CacheImages

-   **Default**: `false`
-   **Options**: `true` or `false`
-   **Description**: Whether to cache images locally.
-   **Details**: If set to `true`, AURA will cache images to reduce load times and improve performance. This is particularly useful for frequently accessed images.Keep in mind that enabling this option will increase disk space usage as images will be stored locally.

---

## SaveImageNextToContent

-   **Default:** `false`
-   **Options:** `true` or `false`
-   **Description:** Whether to save images next to the Media Server content.
-   **Details:**
    -   If `true`, images are saved in the same directory as the Media Server content.
    -   If `false`, images are updated on the Media Server but not saved next to the content.
    -   For **Emby** or **Jellyfin**, this option is ignored (handled by the server).
    -   For **Plex**, this option determines if images are saved next to content.

---

## Logging

-   **Example**:

```yaml
Logging:
    Level: DEBUG
```

### Level

-   **Default**: `TRACE`
-   **Options**: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`
-   **Description**: The logging level for AURA.
-   **Details**:
    -   `TRACE`: Most detailed logging, useful for debugging.
    -   `DEBUG`: Less detailed than TRACE, but still provides useful information for debugging.
    -   `INFO`: General information about the application's operation.
    -   `WARN`: Indicates potential issues that are not necessarily errors.
    -   `ERROR`: Indicates errors that occur during the application's operation.
-   **Note**: The logging level can be adjusted based on your needs. For production environments, it is recommended to use `INFO` or `WARN` to reduce log verbosity. If you run into issues, you can temporarily set it to `DEBUG` or `TRACE` for more detailed logs.

---

## AutoDownload

-   **Example**:

```yaml
AutoDownload:
    Enabled: true
    Cron: "0 0 * * *"
```

### Enabled

-   **Default**: `false`
-   **Options**: `true` or `false`
-   **Description**: Whether to automatically download images from updated sets.
-   **Details**: When downloading images, you have the option to saved sets for "Automatic Downloads". If this option is enabled, AURA will automatically download images from sets that have been updated. This is useful for keeping your media library up-to-date with the latest images without manual intervention.
-   **Note**: Enabling this option may result in increased network usage as AURA will periodically check for updates and download new images.

### Cron

-   **Default**: `0 0 * * *`
-   **Options**: Cron expression
-   **Description**: The cron expression for scheduling automatic downloads.
-   **Details**: This cron expression determines how often AURA checks for updates and downloads images. The default value `0 0 * * *` means that AURA will check for updates every day at midnight. You can modify this expression to change the frequency of automatic downloads according to your needs.
    **Note**: Make sure to use a valid cron expression. You can use online tools like [crontab.guru](https://crontab.guru/) to help you create and validate cron expressions.

---

## Notification

-   **Example**:

```yaml
Notification:
    Provider: Discord
    Webhook: https://discord.com/api/webhooks/your_webhook_url
```

### Provider

-   **Options**: `Discord`
-   **Description**: The notification provider to use for sending notifications.
-   **Details**: Currently, AURA supports Discord as a notification provider. Setting this option is helpful for receiving notifications about the status of automatic downloads and other important events in AURA. If you do not wish to receive notifications, you can leave this option set to `none`.

### Webhook

-   **Description**: The URL of the notifications provider webhook.
-   **Details**: If you choose to use Notification, you must provide the URL of the webhook for the notification provider.

---

## MediaServer

-   **Example**:

```yaml
MediaServer:
    Type: Plex
    URL: http://localhost:32400
    Token: your_token_here
    Libraries:
        - "Movies"
        - "TV Shows"
    SeasonNamingConvention: 2
```

### Type

-   **Options**: `Plex`, `Emby`, `Jellyfin`
-   **Description**: The type of media server you are using.
-   **Details**: This option specifies the type of media server that AURA will interact with. Depending on your choice, AURA will use the appropriate API and methods to manage images and metadata.

### URL

-   **Description**: The URL of the media server.
-   **Details**: This option specifies the URL of the media server that AURA will interact with.
-   **Note**: Make sure to include the protocol (e.g., `http://` or `https://`) in the URL.
-   **Example**: `http://localhost:32400`, `https://my-emby-server.com`, or `http://jellyfin.example.com`.

### Token

-   **Description**: The authentication token for the media server.
-   **Details**: This option specifies the authentication token required to access the media server's API. You can obtain this token from your media server's settings or API documentation.
-   **Note**: The token is necessary for AURA to authenticate and perform actions on your media server. Make sure to keep this token secure and do not share it publicly.

### Libraries

-   **Description**: The name of the media server library to use.
-   **Details**: This option specifies the name of the library on your media server that AURA will interact with. AURA will use this library to manage images and metadata.
-   **Note**: Ensure that the library name matches exactly with the name on your media server, including case sensitivity. Only show and movies libraries are supported.

### SeasonNamingConvention

-   **Default**: `1`
-   **Options**: `1`, `2`
-   **Description**: This is a Plex exclusive requirement. This is the season naming convention for Plex.
-   **Details**:
    -   `1`: Use the format "Season 1" (e.g., "Season 1").
    -   `2`: Use the format "Season 01" (e.g., "Season 01").

---

## Kometa

-   **Example**:

```yaml
Kometa:
    RemoveLabels: true
    Labels:
        - "aura"
        - "kometa"
```

### RemoveLabels

-   **Default**: `false`
-   **Options**: `true` or `false`
-   **Description**: Whether to remove labels from Plex items.
-   **Details**: If set to `true`, AURA will remove labels from Plex items after processing them. This is useful for keeping your media library clean and organized, especially if you use labels for temporary categorization during image processing.

### Labels

-   **Description**: The labels to remove from Plex items.
-   **Details**: This option specifies the labels that AURA will remove from Plex items. Each label should be a new line in the configuration file.

---

## TMDB

-   **Example**:

```yaml
TMDB:
    APIKey: your_tmdb_api_key_here
```

### APIKey

-   **Description**: The API key for The Movie Database (TMDB).
-   **Details**: This option specifies the API key required to access TMDB's API. You can obtain this key by creating an account on [TMDB](https://www.themoviedb.org/) and generating an API key in your account settings.
-   **Note**: This is not yet used for anything, but might be in the future. You can leave this blank for now.

---

## Mediux

-   **Example**:

```yaml
Mediux:
    APIKey: your_mediux_api_key_here
```

### APIKey

-   **Description**: The API key for Mediux.
-   **Details**: This option specifies the API key required to access Mediux's API. This can be obtained by creating an account on [Mediux](https://mediux.io/) and generating an API key in your account settings.
-   **Note**: This is not yet available to the public, but will be in the future.
    If you would like to test out AURA, please contact us on [Discord](https://discord.gg/Sv6wzqfK) to get access to the API key.
