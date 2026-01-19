import { defineComponent, type PropType } from "vue";
import type { Texture } from "@/stores/texture";

export default defineComponent({
  name: "TextureCard",
  props: {
    texture: {
      type: Object as PropType<Texture>,
      required: true,
    },
  },
  setup(props) {
    return () => (
      <div class="texture-card">
        <div class="texture-image">
          <img src={props.texture.path} alt={props.texture.name} />
        </div>
        <div class="texture-info">
          <h3>{props.texture.name}</h3>
          <div class="texture-tags">
            {props.texture.tags.map((tag) => (
              <span key={tag} class="tag">
                {tag}
              </span>
            ))}
          </div>
          <p class="texture-date">{props.texture.createdAt}</p>
        </div>
      </div>
    );
  },
});
