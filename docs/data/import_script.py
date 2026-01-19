#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
智慧工厂系统 - API测试数据导入脚本
用途：向系统批量导入测试数据，用于API测试
使用方法：python import_script.py
"""

import json
import requests
import time

# 配置信息
BASE_URL = "http://localhost:8080"
PRODUCTS_ENDPOINT = BASE_URL + "/api/products"
DATA_FILE = "sample_data.json"

def load_test_data():
    """从文件加载测试数据"""
    try:
        with open(DATA_FILE, 'r', encoding='utf-8') as f:
            return json.load(f)
    except Exception as e:
        print("读取测试数据文件失败: {}".format(e))
        return None

def import_products(products):
    """导入产品数据"""
    print("开始导入产品数据...")
    product_ids = []
    
    for product in products:
        try:
            response = requests.post(PRODUCTS_ENDPOINT, json=product)
            if response.status_code == 200:
                result = response.json()
                if result.get('code') == 0:
                    product_id = result.get('data', {}).get('id')
                    if product_id:
                        product_ids.append(product_id)
                        print("成功导入产品: {} (ID: {})".format(product['name'], product_id))
                    else:
                        print("导入产品成功但无法获取ID: {}".format(product['name']))
                else:
                    print("导入产品失败: {}, 错误: {}".format(product['name'], result.get('message')))
            else:
                print("导入产品请求失败: {}, 状态码: {}".format(product['name'], response.status_code))
        except Exception as e:
            print("导入产品时发生异常: {}".format(e))
    
    return product_ids

def import_components(product_ids, components):
    """为指定产品导入组件数据"""
    print("\n开始导入组件数据...")
    if not product_ids:
        print("没有有效的产品ID，无法导入组件")
        return
    
    # 为每个产品添加组件
    for product_id in product_ids:
        print("\n为产品ID {} 添加组件:".format(product_id))
        for component in components:
            endpoint = "{}/{}/components".format(PRODUCTS_ENDPOINT, product_id)
            try:
                response = requests.post(endpoint, json=component)
                if response.status_code == 200:
                    result = response.json()
                    if result.get('code') == 0:
                        component_id = result.get('data', {}).get('id')
                        print("成功导入组件: {} (ID: {})".format(component['name'], component_id))
                    else:
                        print("导入组件失败: {}, 错误: {}".format(component['name'], result.get('message')))
                else:
                    print("导入组件请求失败: {}, 状态码: {}".format(component['name'], response.status_code))
            except Exception as e:
                print("导入组件时发生异常: {}".format(e))
            
            # 短暂延迟，避免请求过于频繁
            time.sleep(0.5)

def main():
    """主函数"""
    print("智慧工厂系统 - API测试数据导入工具")
    print("=" * 50)
    
    # 加载测试数据
    data = load_test_data()
    if not data:
        print("无法加载测试数据，程序退出")
        return
    
    # 导入产品
    product_ids = import_products(data.get("products", []))
    
    # 导入组件
    import_components(product_ids, data.get("components", []))
    
    print("\n数据导入完成!")
    print("成功导入 {} 个产品及其组件".format(len(product_ids)))
    print("=" * 50)

if __name__ == "__main__":
    main() 