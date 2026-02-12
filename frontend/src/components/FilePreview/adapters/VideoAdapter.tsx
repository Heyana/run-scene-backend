import { defineComponent, ref, onMounted } from "vue";
import { Spin } from "ant-design-vue";
import type { IPreviewAdapter, PreviewAdapterProps } from "../types";

// 视频预览适配器
class VideoPreviewAdapter implements IPreviewAdapter {
  name = "VideoPreviewAdapter";

  private supportedFormats = ["mp4", "webm", "ogg", "avi", "mov"];

  canPreview(format: string): boolean {
    return this.supportedFormats.includes(format.toLowerCase());
  }

  render(props: PreviewAdapterProps) {
    return (
      <VideoPreview
        file={props.file}
        onLoad={props.onLoad}
        onError={props.onError}
      />
    );
  }
}

// 视频预览组件
const VideoPreview = defineComponent({
  name: "VideoPreview",
  props: {
    file: {
      type: Object,
      required: true,
    },
    onLoad: Function,
    onError: Function,
  },
  setup(props) {
    const loading = ref(true);
    const error = ref(false);
    const videoRef = ref<HTMLVideoElement>();

    const handleLoadedData = () => {
      loading.value = false;
      props.onLoad?.();
    };

    const handleError = () => {
      loading.value = false;
      error.value = true;
      props.onError?.(new Error("视频加载失败"));
    };

    onMounted(() => {
      if (videoRef.value) {
        videoRef.value.addEventListener("loadeddata", handleLoadedData);
        videoRef.value.addEventListener("error", handleError);
      }
    });

    return () => (
      <div class="video-preview-container">
        {loading.value && (
          <div class="preview-loading">
            <Spin size="large" tip="加载中..." />
          </div>
        )}

        {error.value ? (
          <div class="preview-error">
            <div class="error-icon">🎬</div>
            <div class="error-text">视频加载失败</div>
            <div class="error-hint">请检查文件格式或网络连接</div>
          </div>
        ) : (
          <video
            ref={videoRef}
            src={props.file.file_url}
            controls
            style={{
              maxWidth: "100%",
              maxHeight: "80vh",
              display: loading.value ? "none" : "block",
            }}
          >
            您的浏览器不支持视频播放
          </video>
        )}
      </div>
    );
  },
});

export default new VideoPreviewAdapter();
