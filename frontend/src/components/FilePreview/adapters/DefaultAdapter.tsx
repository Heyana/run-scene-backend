import { defineComponent } from "vue";
import { Button, Tag } from "ant-design-vue";
import { DownloadOutlined, FileOutlined } from "@ant-design/icons-vue";
import type { IPreviewAdapter, PreviewAdapterProps } from "../types";

// 默认预览适配器（不支持预览的文件类型）
class DefaultPreviewAdapter implements IPreviewAdapter {
  name = "DefaultPreviewAdapter";

  canPreview(format: string): boolean {
    // 默认适配器支持所有格式（作为兜底）
    return true;
  }

  render(props: PreviewAdapterProps) {
    return (
      <DefaultPreview
        file={props.file}
        onLoad={props.onLoad}
        onError={props.onError}
      />
    );
  }
}

// 默认预览组件
const DefaultPreview = defineComponent({
  name: "DefaultPreview",
  props: {
    file: {
      type: Object,
      required: true,
    },
    onLoad: Function,
    onError: Function,
  },
  setup(props) {
    const handleDownload = () => {
      window.open(props.file.file_url, "_blank");
    };

    const formatFileSize = (bytes?: number) => {
      if (!bytes) return "未知";
      const units = ["B", "KB", "MB", "GB"];
      let size = bytes;
      let unitIndex = 0;

      while (size >= 1024 && unitIndex < units.length - 1) {
        size /= 1024;
        unitIndex++;
      }

      return `${size.toFixed(2)} ${units[unitIndex]}`;
    };

    return () => (
      <div class="default-preview-container">
        <div class="preview-content">
          <div class="file-icon">
            <FileOutlined style={{ fontSize: "64px", color: "#999" }} />
          </div>

          <div class="file-info">
            <h3 class="file-name">{props.file.name}</h3>

            <div class="file-meta">
              <Tag color="blue">{props.file.format?.toUpperCase()}</Tag>
              <span class="file-size">
                {formatFileSize(props.file.file_size)}
              </span>
            </div>

            <div class="preview-hint">该文件类型暂不支持在线预览</div>

            <Button
              type="primary"
              size="large"
              icon={<DownloadOutlined />}
              onClick={handleDownload}
              style={{ marginTop: "24px" }}
            >
              下载文件
            </Button>
          </div>
        </div>
      </div>
    );
  },
});

export default new DefaultPreviewAdapter();
