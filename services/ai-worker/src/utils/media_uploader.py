import aiohttp
import asyncio
from typing import Dict, Any, Optional, Tuple
from ..utils.logger import logger

async def download_and_upload_image(
    download_url: str,
    upload_url: str,
    content_type: str = "image/png",
    timeout: int = 60
) -> Tuple[bytes, Dict[str, Any]]:
    """
    Скачивает изображение по одному URL и загружает его по другому URL
    
    Args:
        download_url: URL для скачивания изображения
        upload_url: Предподписанный URL для загрузки
        content_type: MIME-тип изображения
        timeout: Таймаут запроса в секундах
        
    Returns:
        Кортеж (image_data, upload_result), где:
        - image_data: байты скачанного изображения
        - upload_result: словарь с результатом загрузки
    """
    try:
        # Скачиваем изображение
        image_data = None
        async with aiohttp.ClientSession() as session:
            logger.info(f"Downloading image from: {download_url}")
            
            try:
                async with session.get(download_url, timeout=timeout) as response:
                    if response.status != 200:
                        error_text = await response.text()
                        logger.error(f"Failed to download image. Status: {response.status}, Response: {error_text}")
                        return None, {
                            "success": False,
                            "status": response.status,
                            "error": f"Download failed: {error_text}"
                        }
                    
                    image_data = await response.read()
                    logger.info(f"Successfully downloaded image ({len(image_data)} bytes)")
            
            except Exception as e:
                logger.error(f"Error downloading image: {str(e)}")
                return None, {
                    "success": False,
                    "error": f"Download error: {str(e)}"
                }
            
            # Загружаем изображение
            if image_data:
                logger.info(f"Uploading image to: {upload_url}")
                
                try:
                    headers = {"Content-Type": content_type}
                    async with session.put(
                        url=upload_url,
                        data=image_data,
                        headers=headers,
                        timeout=timeout
                    ) as response:
                        if response.status not in (200, 201, 204):
                            error_text = await response.text()
                            logger.error(f"Failed to upload image. Status: {response.status}, Response: {error_text}")
                            return image_data, {
                                "success": False,
                                "status": response.status,
                                "error": f"Upload failed: {error_text}"
                            }
                        
                        logger.info(f"Successfully uploaded image to presigned URL")
                        return image_data, {
                            "success": True,
                            "status": response.status
                        }
                        
                except Exception as e:
                    logger.error(f"Error uploading image: {str(e)}")
                    return image_data, {
                        "success": False,
                        "error": f"Upload error: {str(e)}"
                    }
        
        return None, {"success": False, "error": "Unknown error"}
        
    except Exception as e:
        logger.error(f"Unexpected error in download_and_upload_image: {str(e)}")
        return None, {
            "success": False,
            "error": f"Unexpected error: {str(e)}"
        }