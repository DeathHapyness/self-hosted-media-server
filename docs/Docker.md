# 🐳 Instalação do Docker

## 1. Atualizar o sistema

Em Ubuntu/Debian:

```bash
sudo apt update
sudo apt upgrade -y
```

## 2. Instalar dependências

```bash
sudo apt install -y ca-certificates curl
```

## 3. Adicionar a chave oficial do Docker

```bash
sudo install -m 0755 -d /etc/apt/keyrings

sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  -o /etc/apt/keyrings/docker.asc

sudo chmod a+r /etc/apt/keyrings/docker.asc
```

## 4. Adicionar o repositório

```bash
sudo tee /etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Architectures: $(dpkg --print-architecture)
Signed-By: /etc/apt/keyrings/docker.asc
EOF
```

Atualize:

```bash
sudo apt update
```

## 5. Instalar Docker

```bash
sudo apt install -y \
  docker-ce \
  docker-ce-cli \
  containerd.io \
  docker-buildx-plugin \
  docker-compose-plugin
```

## 6. Testar

```bash
sudo systemctl status docker
```

```bash
sudo docker run hello-world
```

Verifique o Compose:

```bash
docker compose version
```

---

## 👤 Usar Docker sem sudo

Adicione seu usuário ao grupo Docker:

```bash
sudo usermod -aG docker $USER
```

Depois encerre a sessão e entre novamente.

Teste:

```bash
docker ps
```