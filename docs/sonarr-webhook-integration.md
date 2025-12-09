---
layout: default
title: "Sonarr Webhook Integration"
nav_order: 5
description: "Instructions for integrating Sonarr webhooks with aura."
permalink: /sonarr-webhook-integration
---

# Sonarr Webhook Integration

Set up a webhook in Sonarr to notify Aura when an episode file is upgraded. Aura will then automatically redownload the titlecards for upgraded episodes.

---

## Setting Up the Webhook in Sonarr

1. **Open Sonarr** and go to `Settings` → `Connect`.
2. Click the **`+`** button to add a new connection.
3. Select **Webhook** from the connection types.
4. Fill in the following details:

    - **Name:**  
      `Webhook - aura` (or any name you prefer)
    - **Notification Triggers:**  
      Check:
        - `On File Import`
        - `On File Upgrade`
        - `On Rename`
    - **Webhook URL:**
        ```
        http://<AURA_HOST>:<AURA_PORT>/api/sonarr/webhook?library=4K%20Series
        ```
        - Replace `<AURA_HOST>` with the hostname or IP address where Aura is running.
        - Replace `<AURA_PORT>` with the backend port number for Aura. By default, this is `8888` unless you have changed it.
        - Replace `4K%20Series` with your library name (URL encode spaces/special characters).
    - **Method:**  
      `POST`
    - _(Optional)_ **Tags:**  
      Configure if you want to limit the webhook to specific series tags.

5. Click **Test** to ensure the webhook works.
6. Click **Save** to finalize setup.

---

## Verifying the Integration

1. **Upgrade an episode file** in Sonarr.
2. **Check Aura logs** to confirm a request was received from Sonarr.
3. **Verify titlecards** for the upgraded episode are redownloaded in the specified library.
    - _Note:_ Aura waits 15 seconds before processing the webhook to ensure Sonarr and your media server have completed file operations and updated file info.
4. If successful, you should see updated titlecards in your library.

---

🎉 **You have successfully set up Sonarr webhook integration with Aura!**
