---
layout: default
title: "Dockerfile Installation"
nav_order: 2
description: "Instructions for installing and running AURA using a Dockerfile."
parent: Installation
permalink: /installation/docker-file
---

# Dockerfile Installation

To install AURA using a Dockerfile, follow these steps:

1. **Clone the Repository**: Start by cloning the AURA repository from GitHub.

    ```bash
    git clone https://github.com/mediux-team/aura.git
    cd aura
    ```

2. **Create a `config.yml` File**: Before building the Docker image, create a `config.yml` file in the root directory of the project. You can use the [config.yml.sample](https://raw.githubusercontent.com/mediux-team/aura/master/config.yml.sample) as a template.

3. **Build the Docker Image**: Use the following command to build the Docker image:

    ```bash
    docker build -t aura .
    ```

4. Run the Docker Container (adjust the volume paths and ports as needed):

    ```sh
    docker run -d -p 3000:3000 -p 8888:8888 -v '/mnt/user/appdata/aura/':'/config':'rw' -v '/mnt/user/data/media/':'/data/media':'rw' aura
    ```

5. **Access the Web UI**: Open your web browser and go to `http://localhost:3000` to access the AURA web interface.

**Note**: Make sure you have Docker installed on your system before executing these commands. You can check the [official documentation](https://docs.docker.com/get-docker/) for more details.
