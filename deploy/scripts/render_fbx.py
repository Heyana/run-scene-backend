#!/usr/bin/env python3
"""
3D 模型预览图渲染脚本
使用 Blender 命令行渲染 3D 模型的预览图

支持格式: FBX, OBJ
注意: GLB/GLTF 需要 numpy，Blender 3.4.1 存在兼容性问题，暂不支持

使用方法:
    blender -b -P render_fbx.py -- input.fbx output.png [width] [height] [quality]

参数:
    input       - 输入的 3D 模型文件路径 (.fbx, .obj)
    output      - 输出的预览图路径
    width       - 可选，图片宽度，默认 1280
    height      - 可选，图片高度，默认 720
    quality     - 可选，渲染质量: fast(快速,默认), normal(普通), high(高质量)

示例:
    blender -b -P render_fbx.py -- model.fbx preview.png
    blender -b -P render_fbx.py -- model.glb preview.png 1920 1080
    blender -b -P render_fbx.py -- model.obj preview.png 1280 720 high
"""

import bpy
import sys
import os
import math
from mathutils import Vector

def clear_scene():
    """清除默认场景"""
    bpy.ops.wm.read_factory_settings(use_empty=True)

def import_model(file_path):
    """导入 3D 模型文件（支持 FBX, OBJ）"""
    # 获取文件扩展名
    ext = os.path.splitext(file_path)[1].lower()
    
    try:
        if ext == '.fbx':
            bpy.ops.import_scene.fbx(filepath=file_path)
            print(f"✓ 成功导入 FBX: {file_path}")
        elif ext == '.obj':
            bpy.ops.import_scene.obj(filepath=file_path)
            print(f"✓ 成功导入 OBJ: {file_path}")
        else:
            print(f"✗ 不支持的文件格式: {ext}")
            print(f"支持的格式: .fbx, .obj")
            print(f"注意: GLB/GLTF 需要 numpy，当前 Blender 版本存在兼容性问题")
            return False
        return True
    except Exception as e:
        print(f"✗ 导入模型失败: {e}")
        return False

def get_model_bounds():
    """获取模型边界"""
    # 获取所有网格对象
    mesh_objects = [obj for obj in bpy.context.scene.objects if obj.type == 'MESH']
    
    if not mesh_objects:
        return None
    
    # 计算所有对象的边界框（使用世界坐标）
    min_x = min_y = min_z = float('inf')
    max_x = max_y = max_z = float('-inf')
    
    for obj in mesh_objects:
        # 获取对象的世界坐标边界框
        for i in range(8):
            # 将 bound_box 坐标转换为 Vector
            local_coord = Vector(obj.bound_box[i])
            world_coord = obj.matrix_world @ local_coord
            
            min_x = min(min_x, world_coord.x)
            min_y = min(min_y, world_coord.y)
            min_z = min(min_z, world_coord.z)
            max_x = max(max_x, world_coord.x)
            max_y = max(max_y, world_coord.y)
            max_z = max(max_z, world_coord.z)
    
    center_x = (min_x + max_x) / 2
    center_y = (min_y + max_y) / 2
    center_z = (min_z + max_z) / 2
    
    size_x = max_x - min_x
    size_y = max_y - min_y
    size_z = max_z - min_z
    max_size = max(size_x, size_y, size_z)
    
    print(f"✓ 模型边界: 中心({center_x:.2f}, {center_y:.2f}, {center_z:.2f}), 尺寸: {max_size:.2f}")
    
    return {
        'center': (center_x, center_y, center_z),
        'size': max_size,
        'min': (min_x, min_y, min_z),
        'max': (max_x, max_y, max_z)
    }

def setup_camera(bounds):
    """设置相机"""
    if bounds is None:
        # 默认位置
        location = (7, -7, 5)
        rotation = (1.1, 0, 0.785)
    else:
        # 根据模型大小自动调整相机位置
        center = bounds['center']
        size = bounds['size']
        
        # 相机距离 = 模型大小 * 3.5（增加距离以确保完整显示）
        distance = size * 3.5
        
        # 45度角俯视
        angle = math.radians(45)
        
        # 计算相机位置（从右前上方看）
        location = (
            center[0] + distance * 0.7,  # X 轴偏移
            center[1] - distance * 0.7,  # Y 轴偏移
            center[2] + distance * 0.5   # Z 轴高度
        )
        
        # 计算相机朝向（看向模型中心）
        direction = bpy.data.objects.new("Empty", None)
        direction.location = center
        
        rotation = (1.1, 0, 0.785)
    
    bpy.ops.object.camera_add(location=location)
    camera = bpy.context.object
    camera.rotation_euler = rotation
    
    # 如果有边界信息，让相机看向模型中心
    if bounds is not None:
        # 添加 Track To 约束，让相机始终朝向模型中心
        constraint = camera.constraints.new(type='TRACK_TO')
        
        # 创建一个空对象作为目标
        bpy.ops.object.empty_add(location=bounds['center'])
        target = bpy.context.object
        target.name = "CameraTarget"
        
        constraint.target = target
        constraint.track_axis = 'TRACK_NEGATIVE_Z'
        constraint.up_axis = 'UP_Y'
    
    bpy.context.scene.camera = camera
    
    print(f"✓ 相机设置完成: 位置 {location}")
    if bounds:
        print(f"  朝向模型中心: {bounds['center']}")

def setup_lighting(bounds):
    """设置灯光"""
    if bounds is None:
        light_location = (5, 5, 5)
    else:
        center = bounds['center']
        size = bounds['size']
        offset = size * 2
        light_location = (
            center[0] + offset,
            center[1] + offset,
            center[2] + offset
        )
    
    # 主光源（太阳光）
    bpy.ops.object.light_add(type='SUN', location=light_location)
    sun = bpy.context.object
    sun.data.energy = 2.0
    
    # 创建 world 如果不存在
    if bpy.context.scene.world is None:
        world = bpy.data.worlds.new("World")
        bpy.context.scene.world = world
    
    # 补光（环境光）
    world = bpy.context.scene.world
    world.use_nodes = True
    
    # 确保有 Background 节点
    if 'Background' in world.node_tree.nodes:
        bg_node = world.node_tree.nodes['Background']
    else:
        # 创建 Background 节点
        bg_node = world.node_tree.nodes.new('ShaderNodeBackground')
        output_node = world.node_tree.nodes.new('ShaderNodeOutputWorld')
        world.node_tree.links.new(bg_node.outputs[0], output_node.inputs[0])
    
    bg_node.inputs[1].default_value = 0.5  # 环境光强度
    
    print(f"✓ 灯光设置完成")

def setup_render(output_path, width=1280, height=720, quality='fast'):
    """设置渲染参数"""
    scene = bpy.context.scene
    
    # 分辨率
    scene.render.resolution_x = width
    scene.render.resolution_y = height
    scene.render.resolution_percentage = 100
    
    # 输出格式
    scene.render.image_settings.file_format = 'PNG'
    scene.render.image_settings.color_mode = 'RGBA'
    scene.render.filepath = output_path
    
    # 渲染引擎设置（使用 EEVEE 更快）
    scene.render.engine = 'BLENDER_EEVEE'
    
    # 根据质量设置采样数
    if quality == 'high':
        samples = 64
        scene.eevee.use_gtao = True  # 环境光遮蔽
        scene.eevee.use_bloom = True  # 泛光
        scene.eevee.use_ssr = True  # 屏幕空间反射
        quality_desc = "高质量"
    elif quality == 'normal':
        samples = 32
        scene.eevee.use_gtao = True
        scene.eevee.use_bloom = False
        scene.eevee.use_ssr = False
        quality_desc = "普通"
    else:  # fast (默认)
        samples = 16
        scene.eevee.use_gtao = False  # 关闭环境光遮蔽
        scene.eevee.use_bloom = False  # 关闭泛光
        scene.eevee.use_ssr = False  # 关闭屏幕空间反射
        quality_desc = "快速"
    
    scene.eevee.taa_render_samples = samples
    scene.eevee.use_motion_blur = False  # 始终关闭运动模糊
    
    # 背景透明
    scene.render.film_transparent = True
    
    print(f"✓ 渲染设置完成: {width}x{height}, 采样数: {samples} ({quality_desc}模式)")

def render():
    """执行渲染"""
    try:
        bpy.ops.render.render(write_still=True)
        print(f"✓ 渲染完成")
        return True
    except Exception as e:
        print(f"✗ 渲染失败: {e}")
        return False

def main():
    """主函数"""
    # 解析命令行参数
    argv = sys.argv
    
    # 找到 '--' 后面的参数
    if '--' not in argv:
        print("错误: 缺少参数")
        print(__doc__)
        sys.exit(1)
    
    args = argv[argv.index('--') + 1:]
    
    if len(args) < 2:
        print("错误: 至少需要输入文件和输出文件两个参数")
        print(__doc__)
        sys.exit(1)
    
    input_file = args[0]
    output_file = args[1]
    width = int(args[2]) if len(args) > 2 else 1280
    height = int(args[3]) if len(args) > 3 else 720
    quality = args[4] if len(args) > 4 else 'fast'  # 默认快速模式
    
    # 检查输入文件
    if not os.path.exists(input_file):
        print(f"错误: 输入文件不存在: {input_file}")
        sys.exit(1)
    
    # 检查文件格式
    ext = os.path.splitext(input_file)[1].lower()
    if ext not in ['.fbx', '.obj']:
        print(f"错误: 不支持的文件格式: {ext}")
        print("支持的格式: .fbx, .obj")
        print("注意: GLB/GLTF 需要 numpy，当前 Blender 版本存在兼容性问题")
        sys.exit(1)
    
    # 确保输出目录存在
    output_dir = os.path.dirname(output_file)
    if output_dir and not os.path.exists(output_dir):
        os.makedirs(output_dir, exist_ok=True)
    
    print(f"开始渲染 3D 模型预览图...")
    print(f"输入: {input_file}")
    print(f"输出: {output_file}")
    print(f"分辨率: {width}x{height}")
    print(f"质量: {quality}")
    print("-" * 50)
    
    # 执行渲染流程
    clear_scene()
    
    if not import_model(input_file):
        sys.exit(1)
    
    bounds = get_model_bounds()
    setup_camera(bounds)
    setup_lighting(bounds)
    setup_render(output_file, width, height, quality)
    
    if render():
        print("-" * 50)
        print(f"✓ 预览图已保存: {output_file}")
        sys.exit(0)
    else:
        sys.exit(1)

if __name__ == '__main__':
    main()
