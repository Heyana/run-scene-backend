// 适配器注册中心

import type { IPreviewAdapter } from "../types";
import SvgAdapter from "./SvgAdapter";
import ImageAdapter from "./ImageAdapter";
import VideoAdapter from "./VideoAdapter";
import DocumentAdapter from "./DocumentAdapter";
import ModelAdapter from "./ModelAdapter";
import HtmlAdapter from "./HtmlAdapter";
import TextAdapter from "./TextAdapter";
import ExcelAdapter from "./ExcelAdapter";
import WordAdapter from "./WordAdapter";
import DefaultAdapter from "./DefaultAdapter";

// 注册所有适配器（按优先级排序）
const adapters: IPreviewAdapter[] = [
  SvgAdapter, // SVG 优先使用专用适配器
  ImageAdapter,
  VideoAdapter,
  ModelAdapter,
  HtmlAdapter,
  TextAdapter,
  ExcelAdapter,
  WordAdapter,
  DocumentAdapter,
  DefaultAdapter, // 默认适配器放在最后
];

// 根据文件格式获取适配器
export function getAdapter(format: string): IPreviewAdapter {
  const adapter = adapters.find((a) => a.canPreview(format));
  return adapter || DefaultAdapter;
}

// 导出所有适配器
export {
  SvgAdapter,
  ImageAdapter,
  VideoAdapter,
  ModelAdapter,
  HtmlAdapter,
  TextAdapter,
  ExcelAdapter,
  WordAdapter,
  DocumentAdapter,
  DefaultAdapter,
};
