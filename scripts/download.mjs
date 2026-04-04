#!/usr/bin/env node
// download.mjs — Downloads a track from a NetMD device using netmd-exploits.
// Called by md-tui: node scripts/download.mjs <trackIndex> <outputPath>

import { DevicesIds, openNewDevice, NetMDFactoryInterface, getDescriptiveDeviceCode } from 'netmd-js';
import { WebUSB } from 'usb';
import { ExploitStateManager, AtracRecovery, getBestSuited } from 'netmd-exploits';
import fs from 'fs';

const trackIndex = parseInt(process.argv[2], 10);
const outputPath = process.argv[3];

if (isNaN(trackIndex) || !outputPath) {
    console.error('Usage: node download.mjs <trackIndex> <outputPath>');
    process.exit(1);
}

async function main() {
    const webUsb = new WebUSB({ allowedDevices: DevicesIds, deviceTimeout: 10000000 });
    const iface = await openNewDevice(webUsb);
    if (!iface) {
        console.error('ERROR: No NetMD device found');
        process.exit(1);
    }

    try {
        console.error('PROGRESS: connecting');

        // Factory init for MZ-N505
        try { await iface.stop(); } catch(e) {}
        try {
            await iface.sendQuery(new Uint8Array([0x18, 0x09, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00]));
        } catch(e) {}

        const factoryIface = new NetMDFactoryInterface(iface.netMd);

        // Device name query via factory
        try {
            const nameCmd = new Uint8Array([
                0x18, 0x01, 0xff, 0x0e,
                0x4e, 0x65, 0x74, 0x20, 0x4d, 0x44, 0x20, 0x57, 0x61, 0x6c, 0x6b, 0x6d, 0x61, 0x6e
            ]);
            await factoryIface.sendCommand(nameCmd);
            await factoryIface.readReply();
        } catch(e) {}

        // Get device code
        const deviceCode = await factoryIface.getDeviceCode();
        const versionCode = await getDescriptiveDeviceCode(deviceCode);
        console.error(`PROGRESS: device ${versionCode} hwid=${deviceCode.hwid}`);

        // Monkey-patch to return cached values
        factoryIface.getDeviceCode = async () => deviceCode;

        // Also create a proxy for ESM that won't re-init
        console.error('PROGRESS: initializing exploit');

        // Use ESM.create with patched getDeviceCode
        const stateManager = await ExploitStateManager.create(iface, factoryIface);

        const deviceType = { versionCode, hwid: deviceCode.hwid };
        const exploitClass = getBestSuited(AtracRecovery, deviceType);
        const exploit = await stateManager.require(exploitClass);

        console.error('PROGRESS: reading 0');
        const result = await exploit.downloadTrack(trackIndex, (progress) => {
            if (progress && progress.read !== undefined && progress.total !== undefined) {
                const pct = Math.round((progress.read / progress.total) * 100);
                console.error(`PROGRESS: reading ${pct}`);
            }
        });

        // Handle various return types
        let buf;
        if (result instanceof Uint8Array || Buffer.isBuffer(result)) {
            buf = Buffer.from(result);
        } else if (result instanceof ArrayBuffer) {
            buf = Buffer.from(result);
        } else if (result && result.data) {
            buf = Buffer.from(result.data);
        } else if (result && typeof result === 'object') {
            console.error(`DEBUG: result keys=${Object.keys(result)}, constructor=${result.constructor?.name}`);
            // Try to find the audio data in the object
            for (const key of Object.keys(result)) {
                const v = result[key];
                if (v instanceof Uint8Array || Buffer.isBuffer(v)) {
                    buf = Buffer.from(v);
                    console.error(`DEBUG: found data in key '${key}', ${buf.length} bytes`);
                    break;
                }
            }
            if (!buf) {
                throw new Error(`Unexpected result type: ${JSON.stringify(Object.keys(result))}`);
            }
        } else {
            throw new Error(`Unexpected result: ${typeof result}`);
        }

        fs.writeFileSync(outputPath, buf);
        console.error('PROGRESS: done');
        console.log(`OK: ${buf.length} bytes written to ${outputPath}`);

        await stateManager.unload(exploitClass);
    } catch (err) {
        console.error(`ERROR: ${err.stack || err.message}`);
        process.exit(1);
    }
}

main();
