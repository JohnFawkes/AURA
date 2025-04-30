"use server";

import { Vibrant } from "node-vibrant/node";
import path from "path";
import fs from "fs/promises";
import sharp from "sharp";
import colorConvert from "color-convert";
import { cache } from "react";

/**
 * Converts HSL values from Vibrant.js format to a brightened CSS HSL string
 * @param hsl HSL array from Vibrant.js [h, s, l] where h is 0-1, s is 0-1, l is 0-1
 * @param increaseFactor Factor to increase lightness by (0.05 = 5% brighter)
 * @returns CSS HSL string format "hsl(360, 100%, 100%)"
 */
function brightenHsl(
  hsl: [number, number, number],
  increaseFactor = 0.05
): string {
  const h = Math.round(hsl[0] * 360);
  const s = Math.round(hsl[1] * 100);
  // Increase lightness but cap at 95% to avoid pure white
  const l = Math.min(95, Math.round((hsl[2] + increaseFactor) * 100));
  return `hsl(${h}, ${s}%, ${l}%)`;
}

/**
 * Creates an optimized image URL using Next.js Image Optimization API
 * @param imageUrl The original image URL
 * @param width The desired width of the optimized image
 * @param quality The desired quality of the optimized image
 * @returns The optimized image URL
 */
function createOptimizedImageUrl(
  imageUrl: string,
  width = 1920,
  quality = 80
): string {
  // Use DEPLOYED_URL from environment variables or fallback to a default
  const baseUrl =
    process.env.DEPLOYED_URL ||
    (process.env.NODE_ENV === "development"
      ? "http://localhost:5001"
      : "https://mediux.io");

  // Create the Next.js optimized image URL
  return `${baseUrl}/_next/image?url=${encodeURIComponent(imageUrl)}&w=${width}&q=${quality}`;
}

/**
 * Server action to extract colors from an image
 * @param imageUrl URL of the image to extract colors from
 * @param useOptimized Whether to use Next.js image optimization
 * @returns Object containing the extracted primary color
 */
export async function extractColors(
  imageUrl: string | null | undefined,
  useOptimized = true
): Promise<{ primaryColor?: string }> {
  if (!imageUrl) return {};

  try {
    // Create an optimized version of the image URL if requested
    let imagePath = imageUrl;

    if (useOptimized) {
      imagePath = createOptimizedImageUrl(imageUrl);
    }

    // For development environment, still try to access the Next.js cached image
    if (process.env.NODE_ENV === "development") {
      try {
        // Check if we can access the local file system
        const isPotentiallyLocalOrCacheable =
          imageUrl.includes("assets/") ||
          imageUrl.startsWith("https://image.tmdb.org/t/p/");

        if (isPotentiallyLocalOrCacheable) {
          const nextCacheDir = path.join(process.cwd(), ".next/cache/images");

          try {
            const files = await fs.readdir(nextCacheDir);
            const urlParts = imageUrl.split("/");
            const likelyFileNamePart = urlParts[urlParts.length - 1];
            const cachedFile = files.find((file) =>
              file.includes(likelyFileNamePart)
            );

            if (cachedFile) {
              // If we find the cached file, use it instead
              imagePath = path.join(nextCacheDir, cachedFile);
            }
          } catch {
            // Continue with the original or optimized URL in imagePath
          }
        }
      } catch {
        // Continue with the original or optimized URL in imagePath
      }
    }

    // Extract colors using node-vibrant
    const palette = await Vibrant.from(imagePath).getPalette();

    // Extract the primary color - prioritizing LightVibrant
    let primaryColor;
    if (palette.LightVibrant?.hsl) {
      primaryColor = brightenHsl(
        palette.LightVibrant.hsl as [number, number, number]
      );
    }

    return { primaryColor };
  } catch {
    // In case of any error during the process, return empty object
    return {};
  }
}

interface DynamicPalette {
  dynamicLeft: string | null;
  dynamicBottom: string | null;
  vibrant: {
    Vibrant?: string;
    LightVibrant?: string;
    DarkVibrant?: string;
    Muted?: string;
    LightMuted?: string;
    DarkMuted?: string;
  };
}

/**
 * Calculates the perceived brightness of an RGB color
 * @param r - Red component (0-255)
 * @param g - Green component (0-255)
 * @param b - Blue component (0-255)
 * @returns Perceived brightness (0-255)
 */
function calculatePerceivedBrightness(r: number, g: number, b: number): number {
  // Standard formula for perceived brightness
  return Math.sqrt(0.299 * r * r + 0.587 * g * g + 0.114 * b * b);
}

/**
 * Adjusts the color to be more suitable for dark mode while maintaining its character
 * @param hsl - HSL color array [h, s, l]
 * @param rgb - RGB color array [r, g, b]
 * @returns Adjusted HSL color array
 */
function smartAdjustLightness(
  hsl: [number, number, number],
  rgb: [number, number, number]
): [number, number, number] {
  const [h, s, l] = hsl;
  const brightness = calculatePerceivedBrightness(rgb[0], rgb[1], rgb[2]);

  // Only adjust colors that are too bright for dark mode (above 180)
  if (brightness > 180) {
    // Calculate how much to adjust based on how bright it is
    const excessBrightness = brightness - 180;
    const maxExcess = 75; // 255 - 180

    // Calculate adjustment factors
    const brightnessFactor = excessBrightness / maxExcess;

    // Target values for dark mode friendly colors
    const targetLightness = 25; // Lowered from 35 to 25 for more darkening
    const targetSaturation = Math.min(100, s * 1.3); // Increased from 1.2 to 1.3 for more vibrancy

    // Calculate new values with smooth transition
    const newLightness = l - (l - targetLightness) * brightnessFactor;
    const newSaturation = s + (targetSaturation - s) * brightnessFactor;

    return [
      h,
      Math.max(0, Math.min(100, newSaturation)),
      Math.max(0, Math.min(100, newLightness)),
    ];
  }

  return hsl;
}

/**
 * Converts HSL to CSS string
 * @param hsl - HSL color array [h, s, l]
 * @returns CSS HSL string
 */
function hslToCss(hsl: [number, number, number]): string {
  return `hsl(${Math.round(hsl[0])} ${Math.round(hsl[1])}% ${Math.round(hsl[2])}%)`;
}

/**
 * Internal implementation of color extraction and processing
 * This function does the actual work but is wrapped by the cached version below
 */
async function _extractAndProcessColors(
  imageUrl: string
): Promise<DynamicPalette> {
  console.log(`Extracting colors for image: ${imageUrl}`);
  try {
    const response = await fetch(imageUrl, { cache: "no-store" });
    if (!response.ok) {
      throw new Error(`Failed to fetch image: ${response.statusText}`);
    }
    const buffer = Buffer.from(await response.arrayBuffer());

    // Resize to a fixed 16:9 size for consistent processing
    const targetWidth = 1280;
    const targetHeight = 720; // 16:9 aspect ratio
    const imageProcessor = sharp(buffer).resize(targetWidth, targetHeight, {
      fit: "cover",
      position: "center",
    });

    // Define region sizes (10% of width/height)
    const leftWidth = Math.floor(targetWidth * 0.1);
    const bottomHeight = Math.floor(targetHeight * 0.1);
    const bottomTop = targetHeight - bottomHeight;

    // Extract left region
    const leftRegionBuffer = await imageProcessor
      .clone()
      .extract({ left: 0, top: 0, width: leftWidth, height: targetHeight })
      .toBuffer();
    const leftStats = await sharp(leftRegionBuffer).stats();
    const leftR = leftStats.channels[0]?.mean ?? 0;
    const leftG = leftStats.channels[1]?.mean ?? 0;
    const leftB = leftStats.channels[2]?.mean ?? 0;

    // Extract bottom region
    const bottomRegionBuffer = await imageProcessor
      .clone()
      .extract({
        left: 0,
        top: bottomTop,
        width: targetWidth,
        height: bottomHeight,
      })
      .toBuffer();
    const bottomStats = await sharp(bottomRegionBuffer).stats();
    const bottomR = bottomStats.channels[0]?.mean ?? 0;
    const bottomG = bottomStats.channels[1]?.mean ?? 0;
    const bottomB = bottomStats.channels[2]?.mean ?? 0;

    // Convert to HSL
    const leftHsl = colorConvert.rgb.hsl(leftR, leftG, leftB);
    const bottomHsl = colorConvert.rgb.hsl(bottomR, bottomG, bottomB);

    // Use node-vibrant for palette extraction
    const processor = new Vibrant(buffer, {
      // Using low quality to improve performance
      quality: 1,
      // Only include key colors for performance
      colorCount: 64,
    });

    const palette = await processor.getPalette();

    // Create vibrant object with CSS HSL values
    const vibrant: DynamicPalette["vibrant"] = {};
    for (const [key, swatch] of Object.entries(palette)) {
      if (swatch && swatch.hsl && swatch.rgb) {
        // Convert node-vibrant's HSL format to our HSL array
        const hsl: [number, number, number] = [
          swatch.hsl[0] * 360, // Convert 0-1 to 0-360
          swatch.hsl[1] * 100, // Convert 0-1 to 0-100
          swatch.hsl[2] * 100, // Convert 0-1 to 0-100
        ];

        // Apply smart color adjustment for dark mode
        const adjustedHsl = smartAdjustLightness(
          hsl,
          swatch.rgb as [number, number, number]
        );

        // Apply additional brightness to LightMuted color
        if (key === "LightMuted") {
          // Increase lightness by 15% but cap at 95%
          // Also increase saturation by 10% but cap at 90%
          const brightenedHsl: [number, number, number] = [
            adjustedHsl[0],
            Math.min(90, adjustedHsl[1] + 25),
            Math.min(95, adjustedHsl[2] + 18),
          ];
          vibrant[key as keyof typeof vibrant] = hslToCss(brightenedHsl);
        } else {
          vibrant[key as keyof typeof vibrant] = hslToCss(adjustedHsl);
        }
      }
    }

    // Apply the same dark mode adjustments to the left and bottom HSL
    const adjustedLeftHsl = smartAdjustLightness(leftHsl, [
      leftR,
      leftG,
      leftB,
    ] as [number, number, number]);
    const adjustedBottomHsl = smartAdjustLightness(bottomHsl, [
      bottomR,
      bottomG,
      bottomB,
    ] as [number, number, number]);

    const resultPalette: DynamicPalette = {
      dynamicLeft: hslToCss(adjustedLeftHsl),
      dynamicBottom: hslToCss(adjustedBottomHsl),
      vibrant,
    };

    return resultPalette;
  } catch (error) {
    console.error("Error extracting colors:", error);
    return {
      dynamicLeft: null,
      dynamicBottom: null,
      vibrant: {},
    };
  }
}

/**
 * Cached version of extractAndProcessColors
 * Fetches an image, extracts average colors from the left edge and bottom edge,
 * and gets the vibrant palette. Results are cached per URL.
 * @param imageUrl - The URL of the image to process.
 * @returns An object containing the extracted colors and vibrant palette.
 */
export const extractAndProcessColors = cache(
  async (imageUrl: string): Promise<DynamicPalette> => {
    console.log(
      `Color extraction requested for: ${imageUrl} (cached function)`
    );
    return _extractAndProcessColors(imageUrl);
  }
);
