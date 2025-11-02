## [0.9.46] - 2025-11-01

### Added

- Plex Only: Try to dynamically get the season path from episodes on file (falls back to Season 1 or Season 01 - based on Season Naming Conventions)

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
