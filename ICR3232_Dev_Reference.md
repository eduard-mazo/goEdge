# ICR-3232 Advantech — Referencia de Desarrollo y Despliegue

> **Documento de contexto técnico** para desarrollo de aplicaciones embebidas sobre el router industrial Advantech ICR-323x.  
> Versión firmware: `6.6.1 (2026-04-24)` · Plataforma: `ICR_PLATFORM=v3`

---

## 1. Perfil del Dispositivo

| Parámetro | Valor |
|---|---|
| **Modelo** | Advantech ICR-3232 (familia ICR-323x) |
| **OS** | Firmware propietario `ICR-323x 6.6.1` sobre Linux |
| **Kernel** | `Linux 6.1.141` |
| **Arquitectura CPU** | `ARMv7` little-endian hard-float (`armv7-linux-gnueabi`) |
| **Toolchain** | `armv7-linux-gnueabi-gcc (GCC) 7.4.0` |
| **Init system** | BusyBox init (SysV, **no systemd, no procd**) |
| **Shell** | BusyBox ash (`/bin/sh`) |
| **BusyBox** | v1.36.0 |
| **RAM total** | ~503 MB |
| **Docker** | ❌ No disponible, no instalable |
| **Package manager** | ❌ No disponible (`apt`, `opkg`, `apk` ausentes) |

---

## 2. Mapa de Filesystem

### Particiones persistentes

| Dispositivo | Mount point | Tipo | Tamaño | Estado | Uso recomendado |
|---|---|---|---|---|---|
| `/dev/root` | `/` | ext4 rw | ~243 MB | ✅ Persistente (OverlayFS) | Sistema base + `/etc` + `/root` |
| `/dev/mmcblk1p7` | `/opt` | ext4 rw | ~749 MB | ✅ Persistente | Uso del firmware (wadmp_client, python3) |
| `/dev/mmcblk1p6` | `/var/data` | ext4 rw | ~499 MB | ✅ Persistente | Datos de usuario alternativos |

### Particiones volátiles (se borran en reboot)

| Mount point | Tipo | Nota |
|---|---|---|
| `/tmp` | tmpfs | No usar para binarios ni configs |
| `/var` | tmpfs | No usar para logs persistentes |

### ⚠️ Mecanismo OverlayFS

El firmware monta `/` como **overlay** sobre `/mnt/upper` (eMMC). Todo cambio en `/etc`, `/root`, `/usr/local` se persiste en la capa superior (`/mnt/upper`), **no en la imagen base del firmware**. Esto significa:

- Los cambios en `/etc/rc.local`, `/etc/init.d/` **persisten entre reboots**.
- Un factory reset (`/usr/bin/io get rst`) **borra la capa upper** y restaura el firmware base.
- Verificar persistencia de cambios:

```bash
ls /mnt/upper/etc/rc.d/    # archivos modificados en /etc
ls /mnt/upper/root/         # archivos en /root
```

### Ruta recomendada para apps de usuario

```
/root/bin/          ← binarios de aplicación
/root/log/          ← logs persistentes
/root/etc/          ← configuración de la app
/root/scripts/      ← scripts auxiliares
```

```bash
mkdir -p /root/bin /root/log /root/etc
```

---

## 3. Capacidades del Kernel

### cgroups disponibles (v1)

```
cpu, cpuacct, blkio, memory, devices, freezer, net_cls, net_prio, pids, rdma
```

### Overlayfs

Activo — usado por el propio firmware en el boot (`rc.preinit`).

### Módulos de kernel

`lsmod` no disponible en BusyBox. Los módulos disponibles no son consultables directamente.

---

## 4. Secuencia de Boot

```
BusyBox init
    └── /etc/inittab
            └── ::sysinit:/etc/rc.d/rc.preinit
                    ├── Monta /proc
                    ├── Verifica firmware update (fwupdate -c)
                    ├── Verifica reset button
                    ├── Monta OverlayFS (/mnt/upper como upper layer)
                    ├── pivot_root → cambia raíz al overlay
                    └── exec /etc/rc.d/rc.sysinit
                                ├── Inicializa sistema
                                ├── Arranca servicios /etc/rc.d/init.d/
                                └── exec /etc/rc.local  ← PUNTO DE ENTRADA DE USUARIO
```

### Punto de entrada para apps de usuario: `rc.local`

```bash
cat /etc/rc.local
```

```sh
#!/bin/sh
sleep 10                              # espera inicialización de red
python3 /root/modbus_mqtt.py &        # app Python existente
# Agregar apps adicionales aquí
exit 0
```

---

## 5. Ejecutar una App como Proceso Persistente

### Opción A — rc.local directo (simple)

Editar `/etc/rc.local`:

```sh
#!/bin/sh
sleep 10
python3 /root/modbus_mqtt.py &
/root/bin/miapp >> /root/log/miapp.log 2>&1 &
exit 0
```

### Opción B — Watchdog shell (recomendado para producción)

Crear `/root/bin/watchdog_miapp.sh`:

```sh
#!/bin/sh
DAEMON=/root/bin/miapp
LOG=/root/log/miapp.log

while true; do
    if ! ps w | grep -v grep | grep -q miapp; then
        echo "$(date): [RESTART] miapp no encontrado, relanzando..." >> $LOG
        $DAEMON >> $LOG 2>&1 &
    fi
    sleep 30
done
```

En `/etc/rc.local`:

```sh
#!/bin/sh
sleep 10
python3 /root/modbus_mqtt.py &
/root/bin/watchdog_miapp.sh &
exit 0
```

```bash
chmod +x /root/bin/watchdog_miapp.sh
```

### Opción C — init.d script (control start/stop/restart)

Crear `/etc/init.d/miapp`:

> ⚠️ **Sin `start-stop-daemon`.** El BusyBox de este firmware **no** incluye el
> applet `start-stop-daemon` (`start-stop-daemon: not found`). Gestionar el
> proceso con `nohup` + PID file directamente, como abajo. `make service-icr`
> genera exactamente este patrón.

```sh
#!/bin/sh
DAEMON=/root/bin/miapp
RUNDIR=/root/run
PIDFILE=$RUNDIR/miapp.pid
LOG=/root/log/miapp.log

running() {
    [ -f "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null
}

start() {
    if running; then echo "miapp ya corriendo (PID $(cat "$PIDFILE"))"; return 0; fi
    mkdir -p "$RUNDIR" /root/log
    echo "Iniciando miapp..."
    nohup "$DAEMON" >> "$LOG" 2>&1 &
    echo $! > "$PIDFILE"
    echo "OK (PID $(cat "$PIDFILE"))"
}

stop() {
    echo "Deteniendo miapp..."
    if [ -f "$PIDFILE" ]; then
        PID=$(cat "$PIDFILE")
        kill "$PID" 2>/dev/null
        i=0
        while kill -0 "$PID" 2>/dev/null && [ $i -lt 10 ]; do sleep 1; i=$((i+1)); done
        kill -9 "$PID" 2>/dev/null
        rm -f "$PIDFILE"
    fi
    echo "OK"
}

case "$1" in
    start)   start ;;
    stop)    stop  ;;
    restart) stop; sleep 2; start ;;
    status)
        if running; then
            echo "miapp corriendo (PID $(cat "$PIDFILE"))"
        else
            echo "miapp detenido"
        fi
        ;;
    *)
        echo "Uso: $0 {start|stop|restart|status}"
        ;;
esac
```

```bash
chmod +x /etc/init.d/miapp
# Invocar desde rc.local:
# /etc/init.d/miapp start
```

---

## 6. Desarrollo en Go — Cross-Compilación

### Target de compilación

| Parámetro Go | Valor |
|---|---|
| `GOOS` | `linux` |
| `GOARCH` | `arm` |
| `GOARM` | `7` |
| `CGO_ENABLED` | `0` (estático, sin libc externa) |

### Comando de compilación

```bash
GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 \
  go build -ldflags="-s -w" -o miapp ./cmd/miapp
```

- `-s -w` elimina símbolos de debug → binario más pequeño
- `CGO_ENABLED=0` produce binario **100% estático** → sin dependencias en el dispositivo

### Verificar binario antes de copiar

```bash
file miapp
# → ELF 32-bit LSB executable, ARM, EABI5 version 1 (SYSV), statically linked

# Ver tamaño
ls -lh miapp
```

### Manejo de señales en la app (obligatorio para daemon)

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // Log persistente
    f, err := os.OpenFile("/root/log/miapp.log",
        os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err == nil {
        log.SetOutput(f)
        defer f.Close()
    }

    log.Println("Iniciando miapp...")

    // Capturar señales del sistema
    sig := make(chan os.Signal, 1)
    signal.Notify(sig,
        syscall.SIGTERM,  // kill / stop del init
        syscall.SIGINT,   // Ctrl+C
        syscall.SIGHUP,   // reload
    )

    // Lógica principal en goroutine
    go func() {
        // ... tu lógica aquí
    }()

    // Bloquear hasta señal
    s := <-sig
    log.Printf("Señal recibida: %v — shutdown limpio", s)
    // cleanup...
    os.Exit(0)
}
```

---

## 7. Pipeline de Despliegue

### Desde máquina de desarrollo hacia el dispositivo

```bash
# 1. Compilar
GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 \
  go build -ldflags="-s -w" -o miapp ./cmd/miapp

# 2. Copiar al dispositivo
scp miapp root@<IP_DISPOSITIVO>:/root/bin/miapp

# 3. Copiar config si aplica
scp config.yaml root@<IP_DISPOSITIVO>:/root/etc/miapp.yaml

# 4. En el dispositivo: permisos y prueba
ssh root@<IP_DISPOSITIVO> "chmod +x /root/bin/miapp && /root/bin/miapp --version"
```

### Script de deploy automatizado

```bash
#!/bin/bash
# deploy.sh
DEVICE_IP=$1
BINARY=miapp

echo "→ Compilando para ARMv7..."
GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 \
  go build -ldflags="-s -w" -o $BINARY ./cmd/miapp

echo "→ Copiando al dispositivo $DEVICE_IP..."
scp $BINARY root@$DEVICE_IP:/root/bin/$BINARY

echo "→ Reiniciando servicio..."
ssh root@$DEVICE_IP "/etc/init.d/miapp restart"

echo "→ Verificando..."
ssh root@$DEVICE_IP "ps w | grep miapp"
echo "✓ Deploy completado"
```

```bash
chmod +x deploy.sh
./deploy.sh 192.168.1.1
```

---

## 8. Diagnóstico y Logs

### Ver logs de la app

```bash
tail -f /root/log/miapp.log
cat /root/log/miapp.log
```

### Ver procesos corriendo

```bash
ps w
ps w | grep miapp
```

### Ver logs del sistema

```bash
# Kernel y OOM killer
dmesg
dmesg | grep -iE "oom|killed|error"

# Syslog (si está disponible)
cat /var/log/messages 2>/dev/null
cat /var/log/syslog 2>/dev/null

# Auth y sesiones
cat /var/log/auth.log 2>/dev/null
```

### Ver uso de recursos

```bash
free -m                  # RAM
df -Pk                   # Disco
ps w                     # Procesos y memoria por proceso
cat /proc/loadavg        # Carga del sistema
```

### Verificar que rc.local se ejecutó

```bash
# Si tu app arrancó, está en ps
ps w | grep miapp

# Ver log desde boot
cat /root/log/miapp.log | head -5
```

---

## 9. Consideraciones y Restricciones

| Restricción | Detalle |
|---|---|
| **No Docker** | Kernel sin módulos necesarios expuestos, sin package manager |
| **No systemd** | Usar BusyBox init + rc.local o init.d |
| **BusyBox limitado** | Muchas flags de GNU no disponibles (`ps -p`, `cat -A`, `df -h`, `sort`, `lsmod`) |
| **Factory reset** | Borra `/mnt/upper` → elimina todos los cambios en OverlayFS |
| **Watchdog de hardware** | Si el firmware activa el watchdog, la app debe alimentarlo vía `/dev/watchdog` o el sistema reinicia |
| **CGO** | Evitar — no hay libc del host disponible en el dispositivo para linkeo dinámico |
| **`/tmp` y `/var`** | Volátiles — no almacenar binarios, configs ni logs aquí |
| **`/opt`** | Reservado para firmware Advantech (`wadmp_client`, `python3`) — no usar para apps propias |

---

## 10. Referencia Rápida de Comandos BusyBox

| Tarea | Comando BusyBox correcto |
|---|---|
| Ver procesos | `ps w` |
| Ver disco | `df -Pk` |
| Ver RAM | `free -m` |
| Buscar texto | `grep -r "texto" /ruta` |
| Ver archivo paginado | `more /archivo` |
| Editar archivo | `vi /archivo` |
| Ver arquitectura | `cat /proc/version` |
| Ver carga | `cat /proc/loadavg` |
| Ver interfaces de red | `ifconfig` o `ip addr` |
| Matar proceso | `kill <PID>` o `kill -9 <PID>` |
| Proceso en background | `comando &` |
| Ver PID de proceso | `ps w \| grep nombre` |

---

*Generado a partir de sesión de diagnóstico en vivo sobre dispositivo ICR-3232 con firmware 6.6.1 · Junio 2026*
