## [0.9.30] - 2025-10-03

### Fixed

- Updated error details to use structured maps instead of formatted strings for better clarity and consistency

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
