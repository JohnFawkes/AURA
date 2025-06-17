---
layout: default
title: "Unraid Installation"
nav_order: 3
description: "Instructions for installing and running AURA on Unraid."
parent: Installation
permalink: /installation/unraid
---

# Unraid Installation

To install AURA on Unraid, follow these steps:

1. **Create a config.yaml File**:

    - Create a `config.yml` file in a directory on your Unraid server.
    - Use the [config.yml.sample](https://raw.githubusercontent.com/mediux-team/AURA/refs/heads/master/config.yml.sample) file as a reference to create your own configuration.

2. **Search for AURA in Unraid Community Apps**:

    - Open your Unraid web interface.
    - Navigate to the "Apps" tab.
    - Search for "AURA" in the community applications.
    - Select the AURA application by mediux-team.

3. **Configure the AURA Docker Container**:

    - Set the desired configuration options for the AURA container.
    - Make sure to map the necessary ports and volumes.

4. **Start the AURA Container**:

    - Click the "Apply" button to start the AURA container.
    - Wait for the container to start up.

5. **Access the AURA Web UI**:
    - Open your web browser and go to `http://<your-unraid-ip>:3000`.
    - Replace `<your-unraid-ip>` with the IP address of your Unraid server.
