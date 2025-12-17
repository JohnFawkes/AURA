## [0.9.67] - 2025-12-17

### Added

- Revamped Download Modal logic to allow for better handling for failed downloads.
- Added more icons to Download Modal buttons for better visual context.

### Fixed

- Fixed issue where Download Modal would not properly update button text when toggling between "Add to Database Only" and normal download options.
- Fixed hydration issue in Download Modal and Saved Sets Page.
- Fixed issue where Download Modal would not show cursor pointer on Database icon hover.
- Fixed issue where Download Modal button icon would not update properly when toggling between options.

---

## [0.9.66] - 2025-12-16

### Added

- Changed Download Modal to show progress inline with image type for each item.
- Changed Download Modal progress bar to show current step in download process.
- Changed Download Modal to show currently applied sets for better context when downloading images.

### Fixed

- Fixed issue where Download Modal would not show proper error message when download failed.

---

## [0.9.65] - 2025-12-12

### Fixed

- Fixed issue where refreshing on the User Page would cause library section options to not load correctly.

---

## [0.9.64] - 2025-12-11

### Breaking

- SeasonNamingConvention config option has been removed. Please update your configuration accordingly. Read below for more information.

### Added

- Standardized naming convention for image files. This is to help future proof users who want to migrate from Plex to Emby/Jellyfin. The new naming conventions are as follows:
    - Posters: Movie/poster.jpg or Show/poster.jpg
    - Backdrops: Movie/backdrop.jpg or Show/backdrop.jpg
    - Season Posters: Show/seasonXX-poster.jpg
    - Special Season Posters: Show/season-specials-poster.jpg
    - Titlecards: Show/Season #/episode.jpg

    Please note that if you were using SeasonNamingConvention before, this is no longer supported. SeasonNamingConvention was using to determine the Season number but is no longer needed for the new format since all images are saved in root folder. Episode naming conventions remain unchanged. Episode naming for those using "static" will now use the currrent episode numbering format (e.g. S01E01.jpg or S1E1.jpg).

### Fixed

- Fixed issue where running aura in Docker for Windows would cause errors with path names not matching.

---

## [0.9.63] - 2025-12-09

### Fixed

- Fixed issue where response is nil when MediUX returns an error, causing a panic.

---

## [0.9.62] - 2025-12-09

### Fixed

- Fixed issue where Special Season Posters were not being properly named.
- [#120](https://github.com/mediux-team/AURA/issues/120) - Fixed issue where aura would fail to start when external applications were unreachable during startup (e.g. MediUX, Media Server, Sonarr, Radarr).
- Clarify port number for Sonarr in Webhook documentation.

---

## [0.9.61] - 2025-11-18

### Added

- Added Saved Sets to search functionality.
- Added image to Saved Sets table for better visual identification of sets.

### Fixed

- Fixed the layout of the cards so that the badges are aligned properly in cards.
- [#116](https://github.com/mediux-team/AURA/issues/116) - Fixed issue where MediUX token validation would fail if the MediUX service was unreachable, now it properly handles connection errors.
- Improved date formatting to show hours and minutes for recent updates
- Fixed issue where search queries with multiple words would not return expected results.
- Fixed issue where search bar was not being updated when clicking on Database icon in Media Item Page.
- Fixed alignment of Uploader name and avatar in Download Modal for better visual consistency.

---

## [0.9.60] - 2025-11-17

### Added

- Added more breakpoints for larger screens (3xl, 4xl, 5xl, 6xl) to support ultra-wide and 4K/5K/8K monitors.

### Fixed

- Fixed issue with backdrops not working in Emby.
- Fixed issue where multi-set Saved Sets would not save correctly when one was marked for deletion.
- Fixed issue where Selected Types would throw error when undefined in Saved Sets Page.

---

## [0.9.59] - 2025-11-16

### Added

- Added ability to Redownload titlecards when Sonarr episode file is upgrade. This requires you to set up a custom webhook connection from Sonarr to aura. View the [documentation](https://mediux-team.github.io/AURA/sonarr-webhook-integration) for more details.

### Fixed

- Fixed issue where undefined possible action paths would cause error when determining main label for log entries.

---

## [0.9.58] - 2025-11-15

### Added

- Added Bulk Actions to the Saved Sets Page. You can now, force recheck of autodownload, apply tags/labels and delete sets.
- Added total number of sets for each MediUX User in the search results for better context.
- Changed default tab on User Page to "Show Sets" and "Movie Sets" when no Box Sets are available.
- Changed library filter option on User Page to not require unselecting of previous library selection.

### Fixed

- Fixed issue where using the back button would revert page number to 1 instead of the last visited page.

---

## [0.9.57] - 2025-11-14

### Breaking

- The LabelsAndTagsProvider config section has been updated. The field `AddLabelsForSelectedTypes` has been renamed to `AddLabelTagForSelectedTypes` to better reflect its purpose of adding labels in Plex and tags in Sonarr/Radarr. Please update your configuration accordingly. This is a hidden setting (not available in the UI) to add labels/tags for each selected image type during downloads to Plex/Sonarr/Radarr.

### Added

- Added new search bar component to allow searching for Media Items and MediUX Users from anywhere in the app.
- Added Avatar images to the MediUX Users for better visual identification.
- Added some color to the Edit button on the Settings Page for better visibility.
- Added option to add tags in Sonarr/Radarr for each selected image type during downloads. This is a hidden setting (not available in the UI) called `AddLabelTagForSelectedTypes` under the LabelsAndTagsProvider config section.
- Updated changelog format to show which version you are currently on and the changes since your last update.
- Updated Movie Boxsets to use the new Responsive Grid layout for better visibility of items.

### Fixed

- Fixed issue where saving the new config would create a new config.yaml file instead of overwriting existing config.yml.
- Fixed issue where clicking on a Media Item from the Collection Items Page would throw an error for Emby/Jellyfin media servers.
- Fixed issue where images were not being applied correctly for Emby/Jellyfin.
- Fixed issue where Media Item Title on Collection Items Page was causing misalignment in the Carousel.
- Fixed issue where Force Recheck for Auto-Download was not working correctly for large sets.
- Fixed Collections Download modal "Cancel" button to match Main Download modal.
- Cleaned up some inconsistent naming for MediUX in various places.

---

## [0.9.56] - 2025-11-11

### Added

- Added support for Collections in Emby and Jellyfin media servers.
- Changed the Carousel within Collections to be a Responsive Grid for viewing more items at once.

### Fixed

- Fixed issue with Emby/Jellyfin where downloading a Backdrop image would cause it to be uploaded but not selected.

---

## [0.9.55] - 2025-11-11

### Fixed

- Fixed issue where backend didn't pass back info about whether media item exists in database to frontend for download modal.
- Fixed panic error where Sonarr/Radarr didn't find the TMDB ID in the response.
- Fixed issue where login page would not redirect after successful login if user was already authenticated.

---

## [0.9.54] - 2025-11-10

### Added

- Added responsive grid for all pages to improve layout on different screen sizes.
- Added hidden option to add labels to Plex for each selected image type.
- Added support for TMDB Poster and Backdrop URLs in the download queue image selection logic.

### Fixed

- Fixed missing User-Agent header in requests to MediUX GraphQL API.
- Fixed issue where download queue was having panics when Poster for set was empty.
- Fixed issue where pagination was not being reset when number of items per page was changed on another page.

---

## [0.9.53] - 2025-11-09

### Added

- Added a view density slider to allow users to customize the size of Images in the carousel views. This setting is saved in User Preferences.

---

## [0.9.52] - 2025-11-09

### Added

- [#112](https://github.com/mediux-team/AURA/issues/112) - Collections Page will now respect User Preferences for Download Defaults when opening the Collections Download Modal.

### Fixed

- [#113](https://github.com/mediux-team/AURA/issues/113) - Fixed issue where Collections with no media items were being shown on the Collections Page.

---

## [0.9.51] - 2025-11-07

### Added

- Added new route and function to get status of last download in the download queue.

### Fixed

- Fixed issue where download modal was not selecting correct default image type based on user preferences.

---

## [0.9.50] - 2025-11-06

### Added

- Added Collections Page to handle applying Posters and Backdrops to a Collection within Plex. This is a Plex exclusive feature (for now).
- Added aura logo to blank images for better user experience when images are missing.

### Fixed

- Fixed issue with Autodownload not filtering out Seasons/Titlecards that are not present on Media Server.
- Fixed issue where trying to enable "Save Images Locally" during initial onboarding would cause error.
- Fixed issue where Media Server images were no longer being downscaled correctly.
- Fixed issue where Cache Images would cause smaller images to be used even when Mediux.DownloadQuality was set to "original".
- Fixed issue where fetching image from Media Server would not throw an error if the media item did not exist.
- Fixed Documentation to show how to manually edit Config file for MediaServer.Libraries Section.
- Fixed issue where Log Filter actions were not sorted within the Group.

---

## [0.9.49] - 2025-11-02

### Breaking

- If you use Plex as your Media Server, the config option for Season Naming Convention has been moved under the Images -> SaveImagesLocally section. Please update your configuration accordingly. View the [documentation](https://mediux-team.github.io/AURA/config#saveimageslocallyseasonnamingconvention) for more details.

### Added

- Added support for Episode Naming Convention under the Images -> SaveImagesLocally section for Plex Media Servers. This allows you to customize how episode files are named when saving images locally. View the [documentation](https://mediux-team.github.io/AURA/config#saveimageslocallyepisodenamingconvention) for more details.
- Added new Not Found (404) page for better user experience when navigating to invalid routes.
- Added new Error (500) page for better user experience when the application encounters server errors.
- Added Browser Source Mapping for easier debugging of frontend issues.

### Fixed

- Fixed issue where adding first Notification provider would throw an error due to undefined config state

---

## [0.9.48] - 2025-11-02

### Added

- [#108](https://github.com/mediux-team/AURA/issues/108) MediUX images that have a Blurhash will now show a Blurhash placeholder while loading for a better user experience.
- [#110](https://github.com/mediux-team/AURA/issues/110) Added support for custom Webhook notification provider

### Fixed

- [#106](https://github.com/mediux-team/AURA/issues/106) Fixed issue where images for Download Queue were not showing centered correctly on larger screens
- [#107](https://github.com/mediux-team/AURA/issues/107) Fixed issue where season posters were not being downloaded when in Download Queue
- [#109](https://github.com/mediux-team/AURA/issues/109) Fixed issue where large poster sets were not being added to the database.

---

## [0.9.47] - 2025-11-02

### Added

- Added a new Download Queue Page to help manage failed and in-progress downloads. You can redownload items with warnings or errors, and remove files from the queue.
- Added notifications for Download Queue events so you are informed of progress and issues as they happen.
- Added logging details to App Startup to help diagnose issues during initialization.

### Fixed

- Fixed issue with download queue not processing multiple downloads correctly when queued in quick succession.
- Fixed issue with Logs Page not showing up correctly on mobile devices.
- Fixed error handling when Plex doesn't return posters the first time.
- Changed Plex/Sonarr/Radarr label and tag handling to only occur when items are added to the database, instead of during every file download.
- Moved database queue route logic to a separate function for better organization.

---

## [0.9.46] - 2025-11-01

### Added

- Plex Only: Try to dynamically get the season path from episodes on file (falls back to Season 1 or Season 01 - based on Season Naming Conventions)

### Fixed

- Changed exported log file name for clarity

---

## [0.9.45] - 2025-11-01

### Added

- Added 🎉 emojis to changelog headings for better visibility
- Added border to dialog content for better visual distinction
- Updated all filters to use dialog for better visual consistency

---

## [0.9.44] - 2025-11-01

### Added

- Logs Page will now remember filter settings between visits using local storage
- Added backend pagination support for log entries to improve performance on large log files
- Added backend filtering support for log levels, statuses, and actions to enhance log retrieval

### Fixed

- Add image validation and loading state to DimmedBackground component
- Include timestamp in download queue JSON file names to handle multiple sets
- Removed frontend log filtering logic to rely solely on backend filtering for consistency and performance

---

## [0.9.43] - 2025-10-31

### Fixed

- Fixed issue with Plex token not being sent in headers for FetchLibrarySectionOptions and GetMediaServerStatus requests

---

## [0.9.42] - 2025-10-31

### Added

- Standardized Filters on Home Screen to match Saved Sets and Logs page filters
- Combined Sort and Filter into one section for better UX

### Fixed

- Fixed Docker Pulls badge link in README
- Fixed issue on Saved Sets page DB Query where pagination was not working correctly when multiple poster sets exist for a media item
- Fixed issue on Saved Sets page where sort order was not being reset when changed
- Fixed text on Logs Filter drawer description

---

## [0.9.42] - 2025-10-30

### Fixed

- [#104](https://github.com/mediux-team/AURA/issues/104) - Fixed issue with padding on download modal causing misalignment

---

## [0.9.41] - 2025-10-29

### Added

- Added support for adding items to a download queue for better management of multiple downloads

---

## [0.9.40] - 2025-10-27

### Breaking

- Updated log storage format to JSONL, previous log files will not be compatible

### Added

- Revamped the logging system to use JSONL format for better structure and parsing
- Revamped the Logs page to display structured log entries with expandable details
- Added ability to export individual log entries as JSON files
- Logs will also now auto-rotate based on size and age to manage disk space better

---

## [0.9.30] - 2025-10-03

### Added

- Enhance Download Modal to include previous download history for better user tracking
- Can now cycle through poster and season posters on the media item page for series (using touch or mouse drag)
- Added --user support for docker images to run as non-root user
- Added UMASK environment variable support for setting file permissions on downloaded images
- Support for configuring Sonarr/Radarr instances
- Support for setting tags in Sonarr
- Support for setting tags in Radarr
- Added support for testing Sonarr and Radarr connections individually
- Added status indicators for Sonarr and Radarr in settings
- Added support for testing Notification providers individually
- Added status indicators for Notification providers in settings
- Added support for force checking Movies for Rating Key and Path changes on Saved Sets page
- Show already downloaded poster sets at the top of the poster selection carousels

### Fixed

- Updated error details to use structured maps instead of formatted strings for better clarity and consistency
- Reduce size of docker image by switching to a smaller base image and multi-stage builds
- Migrated Database to new structure. TMDB is the main key for Media Items. See breaking changes below.
- Plex labels are now added asynchronously to improve performance when downloading images
- Autodownload will now check for changes to Rating Key (which can change when Media Items are deleted/re-added). This uses TMDB ID as the unique identifier.
- Autodownload will now check for changes to Media Item path (which can change on upgrade or if the file is moved). This is for movies and episodes.
- Fixed issue with clutter on Settings/Onboarding page when item is changed or error on field validation.
- Return Media Item details, Posters and User/Follow Hides in one response to reduce number of API calls

### Breaking

- Database schema has changed. Previous database file should be backed up automatically. All previous entries should be migrated automatically. Any issues should be available in a file called
  `migration_warning_v1.txt` in the same directory as your database.

---

## [0.9.29] - 2025-10-02

### Added

- Change button variant to ghost and update styles for consistency in JumpToTop and RefreshButton components
- [#98](https://github.com/mediux-team/AURA/issues/98) Mask sensitive information in logging for Pushover notifications and media server configuration
- Change home page loading progress bar to be front and center for better visibility on larger libraries
- Changed home page loading sections to use a skeleton loader for better user experience

### Fixed

- [#87](https://github.com/mediux-team/AURA/issues/87) Fixed issue with poster update failing after Plex movie file is replaced
- Adjust footer padding and link text size for improved layout consistency

---

## [0.9.28] - 2025-10-01

### Added

- Added ability to save images locally with configurable path
- Added icons to changelog headings for better visibility
- Added breaking section to changelog for important changes

### Fixed

- Remove cache images from Media Server logic so that images are always fresh
- [#99](https://github.com/mediux-team/AURA/issues/99) - Fixed issue with aura logo in Home Screen icon not having enough padding
- Remove option for Tags from non-Plex media servers
- Remove option for SaveImagesLocally from non-Plex media servers

### Breaking

- If you use SaveImagesNextToContent, please change over to the new SaveImagesLocally option in your config file. The old option has been removed. View the [documentation](https://mediux-team.github.io/AURA/config#saveimageslocallyenabled) for more details.

---

## [0.9.27] - 2025-10-01

### Added

- Added tab navigation for settings and user preferences, enhancing UI organization
- Added release notes dialog to display changelog updates upon new version detection
- Enhance Media Server and MediUX connection with status indicators

### Fixed

- Update search bar to remove on-click animation for improved user experience
- Fixed issue with plex not returning posters
- Update tab triggers in UserSetPage for improved styling and user interaction
- Changed settings cog color to grey for better visual integration
- Update layout and metadata for improved web app manifest and icons
- Add 'select-none' class to SelectTrigger components for improved styling
- Fixed Media Item Page background so that it doesn't move when scrollbar is hidden/shown
- Stop flickering on test connection button by adding an artificial delay

---

## [0.9.26] - 2025-09-25

### Added

- Added GitHub icon next to issue links in changelog for better visibility

### Fixed

- Improve button active state on click for enhanced user experience
- Prevent text selection in badge component for improved usability
- Standardize Destructive button styles across the application
- Rename "Default Image Types" to "Download Defaults" in User Preferences for clarity

---

## [0.9.25] - 2025-09-25

### Added

- [#96](https://github.com/mediux-team/AURA/issues/96) - Added new popup confirmation for destructive actions

### Fixed

- Fixed issue with clearing app cache not working properly on User Preferences Store

---

## [0.9.24] - 2025-09-25

### Added

- Added new changelog page

### Fixed

- Fixed issue with auth token handling

---
