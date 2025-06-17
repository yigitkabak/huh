#!/bin/bash

# Programın yüklü olduğu dizin (Termux için standart)
INSTALL_DIR="/data/data/com.termux/files/usr/bin"
# Binary dosyasının adı
BINARY_NAME="huh"

BINARY_PATH="${INSTALL_DIR}/${BINARY_NAME}"

echo "ℹ️ '${BINARY_NAME}' komutunu kaldırma işlemi başlıyor..."

# Binary dosyasının var olup olmadığını kontrol et
if [ -f "$BINARY_PATH" ]; then
    echo "🗑️ '${BINARY_NAME}' dosyasını siliyorum: ${BINARY_PATH}"
    if rm "$BINARY_PATH"; then
        echo "✅ '${BINARY_NAME}' başarıyla kaldırıldı."
    else
        echo "❌ Hata: '${BINARY_NAME}' dosyasını silerken bir sorun oluştu. İzinleri kontrol edin veya manuel olarak silin: sudo rm ${BINARY_PATH}"
        exit 1
    fi
else
    echo "ℹ️ '${BINARY_NAME}' zaten yüklü değil gibi görünüyor. Kaldırılacak bir şey bulunamadı."
fi

echo "✅ Kaldırma işlemi tamamlandı."
exit 0

