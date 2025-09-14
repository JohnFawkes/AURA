---
layout: default
title: "How to Use"
nav_order: 4
description: "A guide on how to use the MediUX with screenshots and details on various features."
permalink: /how-to-use
has_children: false
has_toc: false
---

# How to Use

---

## Navigation Bar

### Search Bar

On the Home Page, the search bar allows you to find specific Media Items in your library by item name. You can also use the following filters:

-   **Library Name:** `L:Movies:`
-   **Year:** `Y:2023:`
-   **ID:** `ID:12345:`

<sub>_Note: The search bar acts differently on each page._</sub>

---

## Home Page

![Home Page Screenshot](assets/demo/screenshots/home_page.png)
_The main dashboard where you can browse your media library._

**Features:**

-   **Filters:** Filter by library name and processed status.
-   **Sort:** Sort by Date Added or Title (ascending/descending).
-   **Items per Page:** Choose how many media items to display.
-   **Pagination:** Navigate through multiple pages.
-   **Refresh Button:** Reload media items from the server (auto-refreshes after 1 hour of inactivity).

---

## Media Item Page

![Media Item Page Screenshot](assets/demo/screenshots/media_item_page.png)
_Details of a specific media item, including available [Poster Sets](#poster-sets)._

**Features:**

-   See if the item exists in other Libraries or your database.
-   Click Library Name to jump to that page.
-   Click "Already in Database" to go to the [Saved Sets Page](#saved-sets-page) for that item.
-   **Filters:** Toggle hidden uploaders and show/hide Titlecard sets (for shows).
-   **Sort:** Sort Poster Sets by Date Updated or Username.

---

## Poster Sets

![Poster Sets Screenshot](assets/demo/screenshots/poster_set.png)
_Each poster set contains images that can be applied to the media item._

**Features:**

-   Followed uploaders appear at the top.
-   **Movies:** Poster and backdrop images. Some sets are [Collections](#collection-set).
-   **Series:** Poster, backdrop, season posters, and titlecards (latest per season).
-   Dimmed images indicate not in your library.
-   Click Poster Set Name or Magnifying Glass to open [Full Set Page](#full-set-page).
-   Click uploader name to view their [User Page](#user-page).
-   Mix and match images from different sets.

---

### Downloading Images

![Downloading Modal Button Screenshot](assets/demo/screenshots/download_modal_button.png)
_Click the download button to open the modal for downloading images._

**Download Popup:**

-   Select which image types to download.
-   Each Media Item in the set can have its own options.
-   Calculates number of images and download size.
-   Shows progress and errors during download.
-   Displays unique Set ID (click to view on MediUX).
-   After download, view in [Saved Sets Page](#saved-sets-page).

![Download Modal Movie Screenshot](assets/demo/screenshots/download_modal_movie.png)

-   **Movies:** Option for [Add to Database Only](#add-to-database-only).

![Download Modal Show Screenshot](assets/demo/screenshots/download_modal_show.png)

-   **Shows:** Options for [Auto Download](#auto-download) and [Future Updates Only](#future-updates-only).

---

#### Add to Database Only

Save the set to your database without downloading images. Useful for tracking processed movies.

#### Auto Download

Periodically checks for new updates to Poster Sets and downloads them automatically (great for new Season Posters or Titlecards).

#### Future Updates Only

Saves the set to your database and only downloads future updates (helpful if you already have the set).

---

## Full Set Page

![Full Set Page Screenshot 1](assets/demo/screenshots/full_set.png)
_Shows all images in a specific Poster Set, including all Titlecards grouped by season._

-   Click Set Author to go to the [User Page](#user-page).
-   Click "View on MediUX" to see the set on the MediUX site.

---

## User Page

![User Page Screenshot](assets/demo/screenshots/user_page.png)
_View all Poster Sets uploaded by a specific user, including image counts and [Boxsets](#boxset)._

---

## Poster Set Types

-   **Movie Set:** Poster and Backdrop for a specific movie.
-   **Show Set:** Poster, Backdrop, Season Posters, and Titlecards for a show.
-   **Collection Set:** Multiple Posters and Backdrops for a group of movies (e.g., Toy Story 1 & 2).
-   **Boxset:** Multiple Movie Sets, Show Sets, or Collection Sets (e.g., Disney Boxset).

---

## Settings Page

![Settings Page Screenshot](assets/demo/screenshots/settings_page.png)
_View various settings for the application._

---

## Saved Sets Page

![Saved Sets Page Screenshot](assets/demo/screenshots/saved_sets_page.png)
_View all Poster Sets saved to your database file._

**Features:**

-   View and manage all items in your database.
-   Each Media Item appears as a card, with multiple Poster Sets possible.
-   Edit or delete Poster Sets individually.
-   Quickly identify sets with **Auto Download** enabled.

---

### Key Features

-   **Filter:** Easily filter items by Auto Download status.
-   **Sort:** Sort saved sets by Date Added or Title (ascending/descending).

---

### Edit Popup

<div style="text-align:center">
  <img src="assets/demo/screenshots/saved_set_edit.png" alt="Saved Set Edit Popup Screenshot" width="500"/>
</div>

_Customize your saved sets with the Edit Popup:_

-   **Select/Unselect Image Types:** Choose which image types you want (posters, backdrops, titlecards).
-   **Auto Download Toggle (for shows):** Enable/disable automatic downloads for future updates.
-   **Redownload Set:** Instantly redownload the set to refresh or recover images.

> This popup gives you fine-grained control over your saved sets, making it easy to keep your media artwork up to date and tailored to your preferences.

---

## Logs Page

![Logs Page Screenshot](assets/demo/screenshots/logs_page.png)
_View the logs for the application._

**Features:**

-   **Redacted Logs:**  
    View a privacy-friendly version of the logs, so you can safely share them if you need help.

-   **Export Logs:**  
    Download logs for external analysis or reporting.

-   **Clear Logs from Today:**  
    Instantly clear all logs generated today with a single click.

---
