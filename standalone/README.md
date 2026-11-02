# EC2 Instances Viewer

A desktop application that displays AWS EC2 instance information in a GUI table.

## Features

- Select or enter an AWS profile from the screen
- Fetch EC2 instance information via the `awsctrl` command
- Display in table format (ID, Status, Type, PrivateIP, PublicIP, Cost, Name)

## Prerequisites

- Go (module-enabled) — this repository is developed with Go toolchain 1.23 or later.
- The `awsctrl` executable binary (or its path) must be available (specified via the environment variable described below).
- The native GUI uses Gio (`gioui.org`) (no C compiler required).

## Setup

1. If needed, edit the `.env` file and set `AWSCTRL_PATH` to the path of the `awsctrl` executable:

```env
AWSCTRL_PATH=C:\path\to\awsctrl.exe
```

2. Install dependencies:

```bash
cd standalone
go mod tidy
```

3. Run (during development):

```bash
cd standalone
go run .
```

Or double-click `run.bat` to launch it.

## Build

```bash
cd standalone
go build -o ec2viewer.exe .
```

When building on Windows, to prevent a console window (black screen) from
opening alongside the app, build with the `-ldflags="-H=windowsgui"` flag:

```bash
go build -ldflags="-H=windowsgui" -o ec2viewer.exe .
```

## Usage

1. Launch the app
2. Select or enter an AWS profile in the profile field (default: `default`)
3. Press the "Fetch" button to display the list of EC2 instances in the table
4. Click an instance ID to open the Detail panel, then press "Edit SGs" to edit its security groups:
   - **Attached**: check the security groups (from the instance's VPC) that should be attached and press "Apply".
     Instances with more than one network interface can't be changed this way (AWS restriction).
   - **Rules for**: pick an attached security group to see its inbound/outbound rules.
     IPv4 CIDR rules can be deleted ("Delete", then "Confirm?"); IPv6, security-group and prefix-list rules are read-only.
   - **Add rule**: choose ingress/egress and protocol (click to cycle), enter From/To ports
     (ICMP: type/code, empty = all), an IPv4 CIDR and an optional description, then press "Add".
     Security group changes apply to every instance using that group.
   - **New SG**: enter a name (and optionally a description) and press "Create" to create a security group in the
     instance's VPC. With "attach to this instance" checked (default) it is added to the instance's current groups
     and selected, so you can add inbound rules right away. AWS gives a new group an "allow all" outbound rule and no inbound rules.

## Notes


- The clipboard feature uses `github.com/atotto/clipboard`. Depending on the platform, additional tools may be required (e.g., `xclip`/`xsel` on Linux).
- When overwriting the binary on Windows, stop the running process first, or build with a different name (e.g., `-o ec2viewer_alt.exe`).

If there are any other usage steps or screenshots you'd like added, let us know.

