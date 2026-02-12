import { defineComponent, computed, type PropType } from "vue";
import { Modal } from "ant-design-vue";
import { getAdapter } from "./adapters";
import type { FileInfo } from "./types";
import "./style.css";

export default defineComponent({
  name: "FilePreviewModal",
  props: {
    file: {
      type: Object as PropType<FileInfo | null>,
      default: null,
    },
    visible: {
      type: Boolean,
      default: false,
    },
    onClose: {
      type: Function as PropType<() => void>,
      required: true,
    },
  },
  setup(props) {
    const adapter = computed(() => {
      if (!props.file) return null;
      return getAdapter(props.file.format);
    });

    const handleLoad = () => {
      console.log("文件加载成功:", props.file?.name);
    };

    const handleError = (error: Error) => {
      console.error("文件加载失败:", error);
    };

    return () => {
      if (!props.file || !adapter.value) return null;

      return (
        <Modal
          visible={props.visible}
          title={props.file.name}
          footer={null}
          width="90%"
          style={{ maxWidth: "1200px" }}
          onCancel={props.onClose}
          destroyOnClose
          centered
        >
          <div class="file-preview-wrapper">
            {adapter.value.render({
              file: props.file,
              onLoad: handleLoad,
              onError: handleError,
            })}
          </div>
        </Modal>
      );
    };
  },
});
