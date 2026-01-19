#!/usr/bin/env node
/**
 * 智慧工厂系统 - API测试数据导入脚本
 * 用途：向系统批量导入测试数据，用于API测试
 * 使用方法：bun run import_script.js
 */

import fs from 'fs';
import { setTimeout } from 'timers/promises';

// 配置信息
const BASE_URL = "http://localhost:8080";
const PRODUCTS_ENDPOINT = `${BASE_URL}/api/products`;
const DATA_FILE = "./sample_data.json";

/**
 * 从文件加载测试数据
 * @returns {Object|null} 测试数据或null
 */
async function loadTestData() {
    try {
        const data = await fs.promises.readFile(DATA_FILE, 'utf8');
        return JSON.parse(data);
    } catch (error) {
        console.error(`读取测试数据文件失败: ${error.message}`);
        return null;
    }
}

/**
 * 导入产品数据
 * @param {Array} products 产品列表
 * @returns {Array} 成功导入的产品ID列表
 */
async function importProducts(products) {
    console.log("开始导入产品数据...");
    const productIds = [];

    for (const product of products) {
        try {
            const response = await fetch(PRODUCTS_ENDPOINT, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(product),
            });

            if (response.ok) {
                const result = await response.json();
                if (result.code === 0) {
                    const productId = result.data?.id;
                    if (productId) {
                        productIds.push(productId);
                        console.log(`成功导入产品: ${product.name} (ID: ${productId})`);
                    } else {
                        console.log(`导入产品成功但无法获取ID: ${product.name}`);
                    }
                } else {
                    console.log(`导入产品失败: ${product.name}, 错误: ${result.message}`);
                }
            } else {
                console.log(`导入产品请求失败: ${product.name}, 状态码: ${response.status}`);
            }
        } catch (error) {
            console.error(`导入产品时发生异常: ${error.message}`);
        }
    }

    return productIds;
}

/**
 * 为指定产品导入组件数据
 * @param {Array} productIds 产品ID列表
 * @param {Array} components 组件列表
 */
async function importComponents(productIds, components) {
    console.log("\n开始导入组件数据...");
    if (!productIds.length) {
        console.log("没有有效的产品ID，无法导入组件");
        return;
    }

    // 为每个产品添加组件
    for (const productId of productIds) {
        console.log(`\n为产品ID ${productId} 添加组件:`);
        for (const component of components) {
            const endpoint = `${PRODUCTS_ENDPOINT}/${productId}/components`;
            try {
                const response = await fetch(endpoint, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(component),
                });

                if (response.ok) {
                    const result = await response.json();
                    if (result.code === 0) {
                        const componentId = result.data?.id;
                        console.log(`成功导入组件: ${component.name} (ID: ${componentId})`);
                    } else {
                        console.log(`导入组件失败: ${component.name}, 错误: ${result.message}`);
                    }
                } else {
                    console.log(`导入组件请求失败: ${component.name}, 状态码: ${response.status}`);
                }
            } catch (error) {
                console.error(`导入组件时发生异常: ${error.message}`);
            }

            // 短暂延迟，避免请求过于频繁
            await setTimeout(500);
        }
    }
}

/**
 * 主函数
 */
async function main() {
    console.log("智慧工厂系统 - API测试数据导入工具");
    console.log("=".repeat(50));

    // 加载测试数据
    const data = await loadTestData();
    if (!data) {
        console.log("无法加载测试数据，程序退出");
        return;
    }

    // 导入产品
    const productIds = await importProducts(data.products || []);

    // 导入组件
    await importComponents(productIds, data.components || []);

    console.log("\n数据导入完成!");
    console.log(`成功导入 ${productIds.length} 个产品及其组件`);
    console.log("=".repeat(50));
}

// 执行主函数
main().catch(error => {
    console.error(`程序执行出错: ${error.message}`);
    process.exit(1);
}); 