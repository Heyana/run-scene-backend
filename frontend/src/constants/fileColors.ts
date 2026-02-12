// 文件格式颜色配置 - 按类型分类，简化版
export const FILE_CATEGORY_COLORS: Record<string, string> = {
  text: "#1890ff", // 文本 - 蓝色
  video: "#722ed1", // 视频 - 紫色
  spreadsheet: "#52c41a", // 表格 - 绿色
  image: "#fa8c16", // 图片 - 橙色
  document: "#13c2c2", // 文档 - 青色
  model: "#eb2f96", // 3D模型 - 品红色
  web: "#faad14", // 网页 - 金色
  other: "#8c8c8c", // 其他 - 灰色
};

// 根据格式获取分类
export const getFileCategory = (format: string): string => {
  const lowerFormat = format?.toLowerCase() || "";

  const categoryMap: Record<string, string> = {
    // 文本
    txt: "text",
    md: "text",
    log: "text",
    json: "text",
    xml: "text",
    csv: "text",

    // 视频
    mp4: "video",
    avi: "video",
    mov: "video",
    webm: "video",
    mkv: "video",
    flv: "video",

    // 表格
    xlsx: "spreadsheet",
    xls: "spreadsheet",

    // 图片
    png: "image",
    jpg: "image",
    jpeg: "image",
    gif: "image",
    webp: "image",
    svg: "image",
    bmp: "image",
    ico: "image",

    // 文档
    pdf: "document",
    doc: "document",
    docx: "document",
    ppt: "document",
    pptx: "document",

    // 3D模型
    obj: "model",
    glb: "model",
    gltf: "model",
    fbx: "model",

    // 网页
    html: "web",
    htm: "web",
    css: "web",
    js: "web",
    ts: "web",
  };

  return categoryMap[lowerFormat] || "other";
};

// 获取文件格式对应的颜色（按分类）
export const getFileFormatColor = (format: string): string => {
  const category = getFileCategory(format);
  return FILE_CATEGORY_COLORS[category] || "";
};

// 获取分类颜色（别名，保持向后兼容）
export const getCategoryColor = getFileFormatColor;
