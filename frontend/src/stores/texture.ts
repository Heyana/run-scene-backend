import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { apiManager } from "@/api/http";
import type { Texture, TextureListParams } from "@/api/models/texture";

export const useTextureStore = defineStore("texture", () => {
  const textures = ref<Texture[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const total = ref(0);

  const textureCount = computed(() => total.value);

  const fetchTextures = async (params?: TextureListParams) => {
    loading.value = true;
    error.value = null;
    try {
      const response = await apiManager.api.texture.getTextureList(params);
      console.log("Log-- ", response, "response");
      textures.value = response.data.list || [];
      total.value = response.data.total || 0;
    } catch (e) {
      error.value = e instanceof Error ? e.message : "获取材质列表失败";
      textures.value = [];
      total.value = 0;
    } finally {
      loading.value = false;
    }
  };

  const getTextureDetail = async (assetId: string) => {
    loading.value = true;
    error.value = null;
    try {
      const response = await apiManager.api.texture.getTextureDetail(assetId);
      return response.data;
    } catch (e) {
      error.value = e instanceof Error ? e.message : "获取材质详情失败";
      return null;
    } finally {
      loading.value = false;
    }
  };

  const recordUse = async (assetId: string) => {
    try {
      await apiManager.api.texture.recordTextureUse(assetId);
    } catch (e) {
      console.error("记录使用失败:", e);
    }
  };

  const triggerSync = async () => {
    loading.value = true;
    error.value = null;
    try {
      await apiManager.api.texture.triggerSync();
    } catch (e) {
      error.value = e instanceof Error ? e.message : "触发同步失败";
    } finally {
      loading.value = false;
    }
  };

  const getSyncProgress = async () => {
    try {
      const response = await apiManager.api.texture.getSyncProgress();
      return response.data;
    } catch (e) {
      console.error("获取同步进度失败:", e);
      return null;
    }
  };

  return {
    textures,
    loading,
    error,
    total,
    textureCount,
    fetchTextures,
    getTextureDetail,
    recordUse,
    triggerSync,
    getSyncProgress,
  };
});
