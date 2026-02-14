import axios from "axios";
import { ZMessageUtils } from "z_vue3";
import { constant, constApiBasePrefix } from "./const";

export * as apiManager from "./api";

// 响应状态码

const codeMap = {
  // 成功状态码
  0: "成功",
  200: "成功",

  // 客户端错误状态码 (400-499)
  400: "请求参数错误",
  401: "未授权访问",
  403: "禁止操作",
  404: "资源不存在",
  405: "方法不被允许",
  409: "用户已存在",
  422: "请求参数格式错误",
  429: "请求过于频繁",

  // 服务器错误状态码 (500-599)
  500: "服务器内部错误",
  501: "功能未实现",
  502: "网关错误",
  503: "服务不可用",

  // 自定义业务状态码
  410: "用户不存在",
  411: "凭证无效",
  412: "参数无效",
  11000: "记录不存在",
  11001: "产品已存在",
  11002: "组件已存在",
  11003: "记录已存在",
  11004: "记录正在使用中",
};
const isOk = (code: number) => {
  return [0, 200].includes(code);
};
export const http = axios.create({
  baseURL: constApiBasePrefix + "api/",
  //10分钟
  timeout: 10 * 60 * 1000,
});

http.interceptors.request.use(
  (config) => {
    config.headers.Authorization =
      "Bearer " + localStorage.getItem(constant.runSceneBackendToken);
    console.log(
      "Log-- ",
      config.headers.Authorization,
      config,
      "config.headers.Authorization",
    );

    return config;
  },
  (error) => {
    return Promise.reject(error);
  },
);

http.interceptors.response.use(
  (response) => {
    const msg = codeMap[response.data.code as keyof typeof codeMap];
    console.log("Log-- ", msg, response, response.data.msg, "msg");
    if (msg === codeMap[401]) {
      localStorage.removeItem(constant.clientToken);
      localStorage.removeItem("user");

      // 跳转到登录页
      const currentPath = window.location.hash.slice(1); // 移除 # 号
      if (currentPath !== "/login") {
        window.location.hash = `/login?redirect=${encodeURIComponent(currentPath)}`;
      }

      ZMessageUtils.pop(response.data.msg || "未授权访问，请先登录", "danger");
      return Promise.reject(new Error("未授权访问"));
    }
    // return response;
    if (msg && !isOk(response.data.code)) {
      console.log("Log-- ", msg, "msg");
      ZMessageUtils.pop(
        response.data.msg,
        isOk(response.data.code) ? "success" : "danger",
      );
    }
    if (isOk(response.data.code)) {
      return response.data;
    } else {
      return Promise.reject(
        codeMap[response.data.code as keyof typeof codeMap],
      );
    }
  },
  (error) => {
    return Promise.reject(error);
  },
);
