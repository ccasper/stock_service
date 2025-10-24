#!/bin/bash
# Builds the golang binary, archives the source into /opt/$NAME/src, and creates the debian package (all in one).
# How to use:
# 1) Build:
#    ./cmd/stock/build.sh && scp -r cmd/stock/stock_1.0.0.deb gc:
# 3) Install: (Configures watchdog, installs binary, systemd, logrotate, ufw rules)
#    apt-get -y remove stock && dpkg -i /home/<user>/stock_1.0.0.deb || apt-get install -f -y
#
# Or all at once:
#  ./cmd/stock/build.sh && scp -r cmd/stock/stock_1.0.0.deb gc: && ssh gc "sudo apt-get -y remove stock && sudo dpkg -i /home/<user>/stock_1.0.0.deb"
# Note: We create a directory /opt/$NAME/data that's owned by $USER for caching/storing data. apt-get --purge will remove this.
# Systemd logs are: journalctl -u $NAME -f
# Service logs are: tail -f /var/log/$NAME.log

# Get the folder of this build.sh script, which should be the core place we do
# all this work even if we're run from a different location.
MAIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

die() {
    echo >&2 "$@"
    exit 1
}

# Ensure version.go exists
version_file="$MAIN_DIR/version.go"
if [[ ! -f "$version_file" ]]; then
  echo "Error: version file not found at '$version_file'." >&2
  exit 1
fi

# Parse constants from version.go
NAME=$(grep -E '^const[[:space:]]+NAME[[:space:]]*=' "$version_file" | sed -E 's/.*"([^"]+)".*/\1/')
VERSION=$(grep -E '^const[[:space:]]+VERSION[[:space:]]*=' "$version_file" | sed -E 's/.*"([^"]+)".*/\1/')
ROOT=$(grep -E '^const[[:space:]]+ROOT[[:space:]]*=' "$version_file" | sed -E 's/.*"([^"]+)".*/\1/')
SOURCES=$(grep -E '^const[[:space:]]+SOURCES[[:space:]]*=' "$version_file" | sed -E 's/.*"([^"]+)".*/\1/')

if [[ -z "$NAME" || -z "$VERSION" || -z "$SOURCES" ]]; then
  echo "Error: could not parse NAME, VERSION, SOURCES from '$version_file'." >&2
  exit 1
fi


ARCHITECTURE="all" # because we support multiple architectures in one package
MAINTAINER="user <user@gmail.com>"
DESCRIPTION="Runs the appropriate binary for ${NAME} based on system architecture."

# user/group to run this binary on the server.
USER="stock"

# Port to expose this binary on the server.
PORT="8082"

if [[ -z "$NAME" || -z "$VERSION" ]]; then
  echo "Error: could not parse NAME, VERSION from '$version_file'." >&2
  exit 1
fi

BUILD_DIR="$MAIN_DIR/build"
echo "Cleaning up $BUILD_DIR/*"
rm -rf "$BUILD_DIR/"
mkdir -p "$BUILD_DIR"
BIN_DIR="$BUILD_DIR/opt/${NAME}/bin"
mkdir -p "$BIN_DIR"
SERVICE_DIR="$BUILD_DIR/etc/systemd/system"
mkdir -p "$SERVICE_DIR"
DEBIAN_DIR="${BUILD_DIR}/DEBIAN"
mkdir -p "$DEBIAN_DIR"


# First Build the Source Archive
TIMESTAMP=$(date +"%Y-%m-%d-%H-%M-%S")

# Ensure version.go exists
version_file="$MAIN_DIR/version.go"
if [[ ! -f "$version_file" ]]; then
  echo "Error: version file not found at '$version_file'." >&2
  exit 1
fi

echo "Version file: $version_file"
# Parse NAME and VERSION from version.go
NAME=$(grep -E '^const[[:space:]]+NAME[[:space:]]*=' "$version_file" | sed -E 's/.*"([^"]+)".*/\1/')
VERSION=$(grep -E '^const[[:space:]]+VERSION[[:space:]]*=' "$version_file" | sed -E 's/.*"([^"]+)".*/\1/')

if [[ "$ROOT" != /* ]]; then
  # ROOT is relative, make it an absolute path first
  ROOT="$(realpath "${MAIN_DIR}/${ROOT}")"
fi
# Get the relative path of MAIN_DIR from ROOT prefix
REL_MAIN_DIR="${MAIN_DIR#$ROOT}"

# Optionally remove leading slash if it exists
REL_MAIN_DIR="${REL_MAIN_DIR#/}"

echo "Parsed:"
echo "  NAME    = $NAME"
echo "  VERSION = $VERSION"
echo "  SOURCES = $SOURCES"
echo "  ROOT    = $ROOT"
echo "  MAIN_DIR = $MAIN_DIR"
echo "  REL_MAIN_DIR = $REL_MAIN_DIR"

ARCHIVE_NAME="${NAME}-${VERSION}-${TIMESTAMP}.tgz"
DEB_NAME="${NAME}_${VERSION}.deb"


# Step 1: Archive all .go files recursively
# ###############################
echo "Packaging .go files into $ARCHIVE_NAME..."
# Build the list of files to include
mkdir -p "${BUILD_DIR}/opt/${NAME}/src"
INCLUDE_LIST="${BUILD_DIR}/opt/${NAME}/src/include.list"

# Add .go files recursively but prune those in _todelete
PREV_PWD="$(pwd)"
cd "${ROOT}"
for path in $SOURCES; do
  if [[ -e "$path" ]]; then
    echo "  - Adding files from ${ROOT}/$path to $INCLUDE_LIST..."
    find $path -type d -name "_todelete" -prune -o -type f -name "*.go" -print >> "$INCLUDE_LIST"
    find $path -type d -name "_todelete" -prune -o -type f -name "*.sh" -print >> "$INCLUDE_LIST"
    find $path -type d -name "_todelete" -prune -o -type f -name "*.py" -print >> "$INCLUDE_LIST"
    find $path -type d -name "_todelete" -prune -o -type f -name "*.js" -print >> "$INCLUDE_LIST"
    find $path -type d -name "_todelete" -prune -o -type f -name "*.css" -print >> "$INCLUDE_LIST"
    find $path -type d -name "_todelete" -prune -o -type f -name "*.proto" -print >> "$INCLUDE_LIST"
    # Add readme.md files.
    find $path -type d -name "_todelete" -prune -o -type f -iname "readme.md" -print >> "$INCLUDE_LIST"
    # Add root folder go mod files.
    find $path -maxdepth 1 -type f -name "go.*" >> "$INCLUDE_LIST"

  else
    echo "Warning: $path not found" >&2
  fi
done

# Add go mod files in current directory only

# Create tar.gz from the list
mkdir -p "${BUILD_DIR}/opt/${NAME}/src"
tar -czf "${BUILD_DIR}/opt/${NAME}/src/${ARCHIVE_NAME}" -T "$INCLUDE_LIST"

# Create the data directory for the job to write to.
mkdir -p "${BUILD_DIR}/opt/${NAME}/data"
touch "${BUILD_DIR}/opt/${NAME}/data/.storage"


# Build the service
cd "${SCRIPT_DIR}"

env GOARCH=arm64 go build -o "$BIN_DIR/${NAME}-arm64" $MAIN_DIR || die "Unable to create"
env GOARCH=amd64 go build -o "$BIN_DIR/${NAME}-amd64" $MAIN_DIR

# Create runtime wrapper
cat > "${BIN_DIR}/${NAME}" << EOF
#!/bin/bash
VIP=\$(ip -o -4 addr show | awk '{print \$4}' | grep -oE '10\.100\.[0-9]+\.[0-9]+' | head -n 1)
EIP=\$(ip route get 8.8.8.8 | awk '/src/ {print \$7}')

ARCH=\$(uname -m)
if [[ "\$ARCH" == "x86_64" ]]; then
    exec /opt/${NAME}/bin/${NAME}-amd64 --port=$PORT --data=/opt/${NAME}/data
elif [[ "\$ARCH" == "aarch64" ]]; then
    exec /opt/${NAME}/bin/${NAME}-arm64 --port=$PORT --data=/opt/${NAME}/data
else
    echo "Unsupported architecture: \$ARCH"
    exit 1
fi
EOF
chmod +x "${BIN_DIR}/${NAME}"

# Create systemd service
cat > "${SERVICE_DIR}/${NAME}.service" << EOF
[Unit]
Description=Stock Serving Service
After=network.target

[Service]
ExecStart=/opt/${NAME}/bin/${NAME}
Restart=on-failure
# This requires the golang binary to use coreos/go-systemd/daemon to ping systemd.
WatchdogSec=30s
Environment="XDG_CONFIG_HOME=/tmp/.chromium"
Environment="XDG_CACHE_HOME=/tmp/.chromium"
# Set RestartSec to avoid rapid restart loops
RestartSec=5s

# Add StartLimitBurst and StartLimitIntervalSec to control restart frequency
StartLimitBurst=3

# Set WorkingDirectory if your binary expects to run in a specific folder
WorkingDirectory=/opt/${NAME}

# Add TimeoutStartSec and TimeoutStopSec to prevent hangs during start/stop
TimeoutStartSec=30
TimeoutStopSec=30

# Consider adding resource limits for stability/security.
LimitNOFILE=65536
LimitNPROC=512

User=${USER}
Group=${USER}
# Drop capabilities or limit permissions if root is not required.
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full
ProtectHome=yes

StandardOutput=append:/var/log/${NAME}.log
StandardError=append:/var/log/${NAME}.log

[Install]
WantedBy=multi-user.target
EOF

# Create the logrotate configuration
mkdir -p "${BUILD_DIR}/etc/logrotate.d"
cat > "${BUILD_DIR}/etc/logrotate.d/${NAME}" << EOF
/var/log/${NAME}.log {
    daily
    rotate 7
    compress
    missingok
    notifempty
    copytruncate
}
EOF

# runs after install or upgrade
cat > "${DEBIAN_DIR}/postinst" << EOF
#!/bin/bash
set -e

# Add UFW rule
if command -v ufw >/dev/null 2>&1; then
    echo "Allowing ${NAME} service through UFW..."
    ufw allow $PORT/tcp comment "${NAME} service"
    ufw reload
fi

# Create system user if it doesn't exist
if ! id -u "${USER}" >/dev/null 2>&1; then
    echo "Creating system user '${USER}'..."
    useradd --system --shell /usr/sbin/nologin "${USER}"
fi
mkdir -p /home/${USER}
chown -R "${USER}:${USER}" /home/${USER}
chmod -R 775 /home/${USER}

# Set ownership of service files and directories
chown -R root:root /opt/${NAME}
chown -R "${USER}:${USER}" /opt/${NAME}/data

# Enable and start the service
systemctl daemon-reload
systemctl enable ${NAME}.service
systemctl start ${NAME}.service

# Wait up to N seconds for the service to be active
TIMEOUT=30
INTERVAL=1
elapsed=0

PORT=$PORT
check_healthy() {
    echo "Checking curl -sf --connect-timeout 2 --max-time 5 http://localhost:$((PORT + 1))/health >/dev/null"
    curl -sf --connect-timeout 2 --max-time 5 http://localhost:$((PORT + 1))/health >/dev/null
}

echo "Waiting up to \${TIMEOUT} seconds for ${NAME} service to send watchdog ping..."
while true; do
    if check_healthy; then
        echo "$NAME is fully healthy after \${elapsed}s ✅"
        break
    fi

    if [[ "\$ACTIVE" == "failed" || "\$elapsed" -ge "\$TIMEOUT" ]]; then
        echo "$NAME failed to become healthy after \${elapsed}s ❌"
        exit 1
    fi

    sleep "\$INTERVAL"
    elapsed=\$((elapsed + INTERVAL))
done

exit 0
EOF
chmod 755 ${DEBIAN_DIR}/postinst

# runs before removal or upgrade
cat > "${DEBIAN_DIR}/prerm" << EOF
#!/bin/bash
set -e

# Stop and disable the service
systemctl disable ${NAME}.service
systemctl stop ${NAME}.service
systemctl daemon-reload

# The argument is either "purge" or "remove".
if [ "$1" = "purge" ]; then
    echo "Package is being purged (full removal)."
    rm -rf "/opt/${NAME}/data/\*"
    rm -rf "/home/${NAME}/\*
    rm -rf "/home/${NAME}/.\*
    rm -f "/var/log/${NAME}"
    rm -f "/var/log/${NAME}.log\*"
fi

# Delete system user if it exists
if id -u "${USER}" >/dev/null 2>&1; then
    echo "Removing system user '${USER}'..."
    userdel "${USER}" || true
fi

# Remove UFW rule before uninstall
if command -v ufw >/dev/null 2>&1; then
    echo "Removing ${NAME} service rule from UFW..."
    ufw delete allow ${PORT}/tcp
    ufw reload
fi

exit 0
EOF
chmod 755 ${DEBIAN_DIR}/prerm



# Create control file
cat > "${DEBIAN_DIR}/control" << EOF
Package: ${NAME}
Version: ${VERSION}
Section: base
Priority: optional
Depends: chromium | chromium-browser | google-chrome-stable
Architecture: ${ARCHITECTURE}
Maintainer: ${MAINTAINER}
Description: ${DESCRIPTION}
EOF


# Build the package
dpkg-deb --build "$BUILD_DIR" "$MAIN_DIR/$DEB_NAME"

echo "Package built: ${NAME}_${VERSION}.deb"
