import type { AppConfig } from "@/types/config/config";

// Central default
export const defaultAppConfig = (): AppConfig =>
  ({
    auth: {
      enabled: false,
      password: "",
    },
    logging: {
      level: "",
      file: "",
    },
    media_server: {
      type: "",
      url: "",
      api_token: "",
      libraries: [],
      user_id: "",
    },
    mediux: {
      api_token: "",
      download_quality: "",
    },
    auto_download: {
      enabled: false,
      cron: "",
    },
    images: {
      cache_images: { enabled: false },
      save_images_locally: {
        enabled: false,
        path: "",
        episode_naming_convention: "",
      },
    },
    tmdb: {
      api_token: "",
    },
    labels_and_tags: {
      applications: [],
    },
    notifications: {
      enabled: false,
      providers: [],
      templates: {
        app_startup: {
          enabled: true,
          title: "Startup",
          message: "The application has started.",
          include_image: false,
        },
        test_notification: {
          enabled: true,
          title: "Test Notification",
          message: "This is a test notification.",
          include_image: false,
        },
      },
    },
    sonarr_radarr: {
      applications: [{ type: "", library: "", url: "", api_token: "" }],
    },
  }) satisfies AppConfig;
