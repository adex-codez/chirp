import { File, Paths } from "expo-file-system";
import {
  ImageManipulator,
  type ImageRef,
  SaveFormat,
} from "expo-image-manipulator";
import {
  type ImagePickerAsset,
  type ImagePickerResult,
  launchImageLibraryAsync,
} from "expo-image-picker";
import { Platform } from "react-native";

import { apiAuth } from "@/lib/api";
import { parseSessionUser, type SessionUser } from "@/lib/session-user";

const MAX_UPLOAD_BYTES = 2 * 1024 * 1024;
const MAX_SOURCE_BYTES = 32 * 1024 * 1024;
const MAX_SOURCE_DIMENSION = 16384;
const MAX_SOURCE_PIXELS = 64_000_000;
const PICTURE_DIMENSION = 512;

type PictureResult =
  | { kind: "cancelled" }
  | { kind: "saved"; user: SessionUser };

function checkUploadSize(size: number): void {
  if (!Number.isSafeInteger(size) || size <= 0) {
    throw new Error("This picture could not be read. Choose another photo.");
  }
  if (size > MAX_UPLOAD_BYTES) {
    throw new Error("This picture is too large to upload. Choose another photo.");
  }
}

async function normalizePicture(asset: ImagePickerAsset): Promise<string> {
  if (
    !Number.isSafeInteger(asset.width) ||
    !Number.isSafeInteger(asset.height) ||
    asset.width <= 0 ||
    asset.height <= 0
  ) {
    throw new Error("This photo has no usable dimensions. Choose another photo.");
  }
  if (
    asset.width > MAX_SOURCE_DIMENSION ||
    asset.height > MAX_SOURCE_DIMENSION ||
    asset.width * asset.height > MAX_SOURCE_PIXELS
  ) {
    throw new Error("Choose a photo smaller than 64 megapixels.");
  }
  let sourceSize: number | undefined;
  try {
    sourceSize =
      Platform.OS === "web"
        ? asset.file?.size ?? asset.fileSize
        : new File(asset.uri).size;
  } catch {
    throw new Error("This photo could not be read. Choose another photo.");
  }
  if (
    sourceSize !== undefined &&
    (!Number.isSafeInteger(sourceSize) || sourceSize <= 0)
  ) {
    throw new Error("This photo could not be read. Choose another photo.");
  }
  if (sourceSize !== undefined && sourceSize > MAX_SOURCE_BYTES) {
    throw new Error("Choose a photo smaller than 32 MiB.");
  }

  const manipulator = ImageManipulator.manipulate(asset.uri);
  let image: ImageRef | undefined;
  try {
    const side = Math.min(asset.width, asset.height);
    if (asset.width !== asset.height) {
      manipulator.crop({
        originX: Math.floor((asset.width - side) / 2),
        originY: Math.floor((asset.height - side) / 2),
        width: side,
        height: side,
      });
    }
    if (side > PICTURE_DIMENSION) {
      manipulator.resize({ width: PICTURE_DIMENSION, height: PICTURE_DIMENSION });
    }
    image = await manipulator.renderAsync();
    const output = await image.saveAsync({ format: SaveFormat.JPEG, compress: 0.85 });
    return output.uri;
  } catch {
    throw new Error("This photo could not be prepared. Choose another photo.");
  } finally {
    for (const reference of [image, manipulator]) {
      try {
        if (
          Platform.OS === "web" &&
          reference &&
          "uri" in reference &&
          typeof reference.uri === "string" &&
          reference.uri.startsWith("blob:")
        ) {
          URL.revokeObjectURL(reference.uri);
        }
        reference?.release();
      } catch (error) {
        console.warn("profile_picture_reference_cleanup_failed", { error });
      }
    }
  }
}

async function readPicture(uri: string, signal: AbortSignal): Promise<ArrayBuffer> {
  if (Platform.OS === "web") {
    const response = await fetch(uri, { signal });
    if (!response.ok) {
      throw new Error("This picture could not be read. Choose another photo.");
    }
    const blob = await response.blob();
    checkUploadSize(blob.size);
    return blob.arrayBuffer();
  }
  const file = new File(uri);
  checkUploadSize(file.size);
  return file.arrayBuffer();
}

export async function chooseAndUploadPicture({
  signal,
}: {
  signal: AbortSignal;
}): Promise<PictureResult> {
  if (signal.aborted) return { kind: "cancelled" };
  let selection: ImagePickerResult;
  try {
    selection = await launchImageLibraryAsync({
      mediaTypes: ["images"],
      allowsEditing: true,
      aspect: [1, 1],
      quality: 1,
      allowsMultipleSelection: false,
    });
  } catch {
    if (signal.aborted) return { kind: "cancelled" };
    throw new Error("Could not open your photos. Check photo access in Settings and try again.");
  }
  if (signal.aborted || selection.canceled) return { kind: "cancelled" };
  const asset = selection.assets[0];
  if (!asset) throw new Error("No photo was selected. Choose a photo again.");

  let outputURI: string | undefined;
  try {
    outputURI = await normalizePicture(asset);
    if (signal.aborted) return { kind: "cancelled" };
    const bytes = await readPicture(outputURI, signal);
    checkUploadSize(bytes.byteLength);
    if (signal.aborted) return { kind: "cancelled" };
    const data = await apiAuth<unknown>("/auth/profile-picture", {
      method: "POST",
      headers: { "Content-Type": "image/jpeg" },
      body: bytes,
      signal,
    });
    return { kind: "saved", user: parseSessionUser(data) };
  } finally {
    if (outputURI && outputURI !== asset.uri) {
      try {
        if (Platform.OS === "web") {
          if (outputURI.startsWith("blob:")) URL.revokeObjectURL(outputURI);
        } else {
          const cachePrefix = Paths.cache.uri.replace(/\/?$/, "/");
          if (outputURI.startsWith(cachePrefix)) {
            const file = new File(outputURI);
            if (file.exists) file.delete();
          }
        }
      } catch (error) {
        console.warn("profile_picture_file_cleanup_failed", { error });
      }
    }
  }
}
