#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Starting 'huh' command line tool installation...${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}Go derleyicisi bulunamadı.${NC}"
    echo "Lütfen önce Go'yu kurun. Debian/Ubuntu için şunu kullanabilirsiniz: sudo apt update && sudo apt install golang"
    exit 1
fi

echo "Go derleyicisi bulundu."

PROJECT_ROOT_DIR="$(dirname "$(pwd)")"
PROJECT_ROOT_DIR="$(dirname "${PROJECT_ROOT_DIR}")"

CURRENT_DIR=$(pwd)

echo "Go modülü işlemleri için proje kök dizinine gidiliyor: ${PROJECT_ROOT_DIR}"
cd "${PROJECT_ROOT_DIR}" || { echo -e "${RED}Proje kök dizinine geçilemedi.${NC}"; exit 1; }

if [ ! -f "go.mod" ]; then
    echo "go.mod dosyası bulunamadı. Go modülü başlatılıyor..."
    go mod init huh-cli || { echo -e "${RED}Go modülü başlatılamadı.${NC}"; exit 1; }
    echo "Go modülü başarıyla başlatıldı."
fi

echo "Go bağımlılıkları indiriliyor ve düzenleniyor..."
go mod tidy || { echo -e "${RED}Go bağımlılıkları indirilirken bir hata oluştu.${NC}"; exit 1; }
echo "Go bağımlılıkları hazır."

echo "Kurulum betiği dizinine geri dönülüyor: ${CURRENT_DIR}"
cd "${CURRENT_DIR}" || { echo -e "${RED}Kurulum betiği dizinine geri dönülemedi.${NC}"; exit 1; }

echo " 'huh' ikili dosyası derleniyor..."
go build -o huh ../main.go

if [ ! -f "huh" ]; then
    echo -e "${RED}Derleme başarısız oldu. Lütfen yukarıdaki hataları kontrol edin.${NC}"
    exit 1
fi

echo "Derleme başarılı."

INSTALL_DIR="/usr/local/bin"
echo " 'huh' ${INSTALL_DIR} dizinine kurulmaya çalışılıyor..."

if mv huh "${INSTALL_DIR}/huh"; then
    echo -e "${GREEN}✅ 'huh' başarıyla kuruldu!${NC}"
    echo "Artık 'huh' komutunu terminalinizin herhangi bir yerinde kullanabilirsiniz."
    echo "Deneyin: huh help"
else
    echo -e "${RED}'huh' dosyası ${INSTALL_DIR} dizinine taşınamadı.${NC}"
    echo "Lütfen bu betiği 'sudo' ile çalıştırmayı deneyin:"
    echo "sudo bash linux.sh"
    exit 1
fi
