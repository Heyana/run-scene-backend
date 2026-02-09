// 环境检测
const isDev =
  import.meta.env.MODE === "development" ||
  process.env.NODE_ENV === "development";
const mode = import.meta.env.MODE || "production";

// API 基础地址配置
const pubUrl = "http://192.168.3.10:23359/";
// API 基础地址配置
const getBaseUrl = () => {
  // return pubUrl;
  // 默认配置
  return isDev ? "http://192.168.3.39:23359/" : pubUrl;
};
const base = getBaseUrl();
export const constApiBasePrefix = base;

export const constApi = {
  imgPrefix: base + "static/static/",
  fillImg: (img: string) => {
    return constApi.imgPrefix + img;
  },
};

export const constant = {
  token: "enterprise-token",
  clientToken: "client-token",
  isDev,
  mode,
  apiBase: base,
};

// 导出环境相关信息
export const env = {
  MODE: mode,
  DEV: isDev,
  PROD: !isDev,
  API_BASE_URL: base,
  APP_TITLE: import.meta.env.VITE_APP_TITLE || "企业管理系统",
  APP_DEBUG: import.meta.env.VITE_APP_DEBUG === "true" || isDev,
};

export const constPop = {
  enterpriseTip: "enterpriseTip",
  customEDetail: "customEDetail",
  sectionPop: "sectionPop",
  ai: "ai",
  videoPop: "videoPop",
  imagePop: "imagePop",
};
