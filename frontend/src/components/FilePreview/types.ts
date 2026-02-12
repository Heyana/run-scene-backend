// 文件预览类型定义

import type { JSX } from "vue/jsx-runtime";

export interface FileInfo {
  id: number;
  name: string;
  format: string;
  file_url: string;
  thumbnail_url?: string;
  file_size?: number;
  type?: string;
}

export interface PreviewProps {
  file: FileInfo;
  visible: boolean;
  onClose: () => void;
}

export interface PreviewAdapterProps {
  file: FileInfo;
  onLoad?: () => void;
  onError?: (error: Error) => void;
}

// 预览适配器接口
export interface IPreviewAdapter {
  // 判断是否支持该文件格式
  canPreview(format: string): boolean;

  // 渲染预览内容
  render(props: PreviewAdapterProps): JSX.Element;

  // 适配器名称
  name: string;
}

// 文件类型分类
export enum FileCategory {
  IMAGE = "image",
  VIDEO = "video",
  DOCUMENT = "document",
  AUDIO = "audio",
  ARCHIVE = "archive",
  CODE = "code",
  OTHER = "other",
}

// 格式映射
export const FORMAT_CATEGORY_MAP: Record<string, FileCategory> = {
  // 图片
  jpg: FileCategory.IMAGE,
  jpeg: FileCategory.IMAGE,
  png: FileCategory.IMAGE,
  gif: FileCategory.IMAGE,
  webp: FileCategory.IMAGE,
  bmp: FileCategory.IMAGE,
  svg: FileCategory.IMAGE,
  ico: FileCategory.IMAGE,

  // 视频
  mp4: FileCategory.VIDEO,
  webm: FileCategory.VIDEO,
  ogv: FileCategory.VIDEO,
  avi: FileCategory.VIDEO,
  mov: FileCategory.VIDEO,
  mkv: FileCategory.VIDEO,
  flv: FileCategory.VIDEO,

  // 文档
  pdf: FileCategory.DOCUMENT,
  doc: FileCategory.DOCUMENT,
  docx: FileCategory.DOCUMENT,
  xls: FileCategory.DOCUMENT,
  xlsx: FileCategory.DOCUMENT,
  ppt: FileCategory.DOCUMENT,
  pptx: FileCategory.DOCUMENT,
  txt: FileCategory.DOCUMENT,
  md: FileCategory.DOCUMENT,

  // 音频
  mp3: FileCategory.AUDIO,
  wav: FileCategory.AUDIO,
  ogg: FileCategory.AUDIO,
  flac: FileCategory.AUDIO,

  // 压缩包
  zip: FileCategory.ARCHIVE,
  rar: FileCategory.ARCHIVE,
  "7z": FileCategory.ARCHIVE,
  tar: FileCategory.ARCHIVE,
  gz: FileCategory.ARCHIVE,

  // 代码
  js: FileCategory.CODE,
  ts: FileCategory.CODE,
  jsx: FileCategory.CODE,
  tsx: FileCategory.CODE,
  vue: FileCategory.CODE,
  css: FileCategory.CODE,
  scss: FileCategory.CODE,
  html: FileCategory.CODE,
  json: FileCategory.CODE,
  xml: FileCategory.CODE,
};

export function getFileCategory(format: string): FileCategory {
  return FORMAT_CATEGORY_MAP[format.toLowerCase()] || FileCategory.OTHER;
}
