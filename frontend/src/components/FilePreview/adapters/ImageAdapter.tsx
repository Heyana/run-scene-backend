import { defineComponent, ref } from "vue";
import { Image, Spin } from "ant-design-vue";
import type { IPreviewAdapter, PreviewAdapterProps } from "../types";

// 图片预览适配器
class ImagePreviewAdapter implements IPreviewAdapter {
  name = "ImagePreviewAdapter";

  private supportedFormats = [
    "jpg",
    "jpeg",
    "png",
    "gif",
    "webp",
    "bmp",
    "svg",
    "ico",
  ];

  canPreview(format: string): boolean {
    return this.supportedFormats.includes(format.toLowerCase());
  }

  render(props: PreviewAdapterProps) {
    return (
      <ImagePreview
        file={props.file}
        onLoad={props.onLoad}
        onError={props.onError}
      />
    );
  }
}

// 图片预览组件
const ImagePreview = defineComponent({
  name: "ImagePreview",
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

    const handleLoad = () => {
      loading.value = false;
      props.onLoad?.();
    };

    const handleError = (e: Event) => {
      loading.value = false;
      error.value = true;
      props.onError?.(new Error("图片加载失败"));
    };

    return () => (
      <div class="image-preview-container">
        {loading.value && (
          <div class="preview-loading">
            <Spin size="large" tip="加载中..." />
          </div>
        )}

        {error.value ? (
          <div class="preview-error">
            <div class="error-icon">📷</div>
            <div class="error-text">图片加载失败</div>
          </div>
        ) : (
          <div style={{ display: loading.value ? "none" : "block" }}>
            <Image
              src={props.file.file_url}
              alt={props.file.name}
              style={{ maxWidth: "100%", maxHeight: "80vh" }}
              preview={{
                src: props.file.file_url,
              }}
              onLoad={handleLoad}
              onError={handleError}
            />
          </div>
        )}
      </div>
    );
  },
});

export default new ImagePreviewAdapter();
