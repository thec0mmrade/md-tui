#!/usr/bin/env node
// firmware-dump.mjs — Dumps firmware ROM + SRAM using netmd-exploits FirmwareDumper

import { DevicesIds, openNewDevice, NetMDFactoryInterface, getDescriptiveDeviceCode } from 'netmd-js';
import { WebUSB } from 'usb';
import { ExploitStateManager, FirmwareDumper } from 'netmd-exploits';
import fs from 'fs';

const outputPath = process.argv[2] || 'firmware.bin';

async function main() {
    const webUsb = new WebUSB({ allowedDevices: DevicesIds, deviceTimeout: 10000000 });
    const iface = await openNewDevice(webUsb);
    if (!iface) {
        console.error('ERROR: No NetMD device found');
        process.exit(1);
    }

    try {
        console.error('PROGRESS: connecting');

        // Factory init
        try { await iface.stop(); } catch(e) {}
        try {
            await iface.sendQuery(new Uint8Array([0x18, 0x09, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00]));
        } catch(e) {}

        const factoryIface = new NetMDFactoryInterface(iface.netMd);

        try {
            const nameCmd = new Uint8Array([
                0x18, 0x01, 0xff, 0x0e,
                0x4e, 0x65, 0x74, 0x20, 0x4d, 0x44, 0x20, 0x57, 0x61, 0x6c, 0x6b, 0x6d, 0x61, 0x6e
            ]);
            await factoryIface.sendCommand(nameCmd);
            await factoryIface.readReply();
        } catch(e) {}

        const deviceCode = await factoryIface.getDeviceCode();
        const versionCode = await getDescriptiveDeviceCode(deviceCode);
        console.error(`PROGRESS: device ${versionCode} hwid=${deviceCode.hwid}`);

        factoryIface.getDeviceCode = async () => deviceCode;

        console.error('PROGRESS: creating exploit state manager');
        const stateManager = await ExploitStateManager.create(iface, factoryIface);

        console.error('PROGRESS: requiring FirmwareDumper');
        const dumper = await stateManager.require(FirmwareDumper);

        console.error('PROGRESS: reading firmware (this takes ~10 minutes)...');
        const { rom, ram } = await dumper.readFirmware((progress) => {
            console.error(`PROGRESS: ${JSON.stringify(progress)}`);
        });

        // Write ROM
        fs.writeFileSync(outputPath, Buffer.from(rom));
        console.error(`ROM: ${rom.length} bytes -> ${outputPath}`);

        // Write SRAM
        const sramPath = outputPath.replace(/\.[^.]+$/, '.sram');
        fs.writeFileSync(sramPath, Buffer.from(ram));
        console.error(`SRAM: ${ram.length} bytes -> ${sramPath}`);

        console.log(`OK: ROM=${rom.length} SRAM=${ram.length}`);
        process.exit(0);
    } catch (err) {
        console.error(`ERROR: ${err.stack || err.message}`);
        process.exit(1);
    }
}

main();
