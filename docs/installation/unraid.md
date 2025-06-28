---
layout: default
title: "Unraid Installation"
nav_order: 3
description: "Instructions for installing and running aura on Unraid."
parent: Installation
permalink: /installation/unraid
---

# Unraid Installation

To install aura on Unraid, follow these steps:

1. **Create a config.yaml File**:

    - Create a `config.yml` file in a directory on your Unraid server.
    - Use the [config.yml.sample](https://raw.githubusercontent.com/mediux-team/AURA/refs/heads/master/config.yml.sample) file as a reference to create your own configuration.

2. **Search for aura in Unraid Community Apps**:

    - Open your Unraid web interface.
    - Navigate to the "Apps" tab.
    - Search for "aura" in the community applications.
    - Select the aura application by mediux-team.

3. **Configure the aura Docker Container**:

    - Set the desired configuration options for the aura container.
    - Make sure to map the necessary ports and volumes.

4. **Start the aura Container**:

    - Click the "Apply" button to start the aura container.
    - Wait for the container to start up.

5. **Access the aura Web UI**:
    - Open your web browser and go to `http://<your-unraid-ip>:3000`.
    - Replace `<your-unraid-ip>` with the IP address of your Unraid server.
