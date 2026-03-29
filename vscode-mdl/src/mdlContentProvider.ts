// SPDX-License-Identifier: Apache-2.0

import * as vscode from 'vscode';
import * as cp from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import { resolvedMxcliPath } from './extension';

interface CachedFile {
	content: Uint8Array;
	mtime: number;
	ctime: number;
}

/**
 * FileSystemProvider for mendix-mdl:// URIs.
 *
 * - readFile:  runs `mxcli describe` to produce MDL source
 * - writeFile: writes content to a temp file, runs `mxcli exec` to apply changes
 *
 * Documents are editable; Ctrl+S saves changes back to the Mendix project.
 */
export class MdlFileSystemProvider implements vscode.FileSystemProvider {
	private mxcliPath: string;
	private mprPath: string | undefined;
	private cache = new Map<string, CachedFile>();

	private _emitter = new vscode.EventEmitter<vscode.FileChangeEvent[]>();
	readonly onDidChangeFile = this._emitter.event;

	constructor() {
		const config = vscode.workspace.getConfiguration('mdl');
		this.mxcliPath = resolvedMxcliPath();
		const configured = config.get<string>('mprPath', '');
		this.mprPath = configured || undefined;
	}

	updateConfig(): void {
		const config = vscode.workspace.getConfiguration('mdl');
		this.mxcliPath = resolvedMxcliPath();
		const configured = config.get<string>('mprPath', '');
		this.mprPath = configured || undefined;
	}

	watch(): vscode.Disposable {
		return new vscode.Disposable(() => {});
	}

	stat(uri: vscode.Uri): vscode.FileStat {
		const cached = this.cache.get(uri.toString());
		const now = Date.now();
		return {
			type: vscode.FileType.File,
			ctime: cached?.ctime ?? now,
			mtime: cached?.mtime ?? now,
			size: cached?.content.byteLength ?? 0,
		};
	}

	readDirectory(): [string, vscode.FileType][] {
		return [];
	}

	createDirectory(): void {
		throw vscode.FileSystemError.NoPermissions('Cannot create directories');
	}

	async readFile(uri: vscode.Uri): Promise<Uint8Array> {
		// URI format: mendix-mdl://describe/<type>/<qualifiedName>
		const parts = uri.path.split('/').filter(Boolean);
		if (parts.length < 2 || uri.authority !== 'describe') {
			throw vscode.FileSystemError.FileNotFound('Invalid URI format');
		}

		const elementType = parts[0];
		const qualifiedName = parts.slice(1).join('/');

		const mprFile = await this.findMprFile();
		if (!mprFile) {
			throw vscode.FileSystemError.FileNotFound('No .mpr file found. Set mdl.mprPath in settings.');
		}

		const content = await this.runDescribe(mprFile, elementType, qualifiedName);
		if (content.startsWith('-- Error describing')) {
			throw vscode.FileSystemError.FileNotFound(content);
		}

		const encoded = new TextEncoder().encode(content);
		const now = Date.now();
		this.cache.set(uri.toString(), { content: encoded, mtime: now, ctime: now });
		return encoded;
	}

	async writeFile(uri: vscode.Uri, content: Uint8Array): Promise<void> {
		const mprFile = await this.findMprFile();
		if (!mprFile) {
			vscode.window.showErrorMessage('No .mpr file found. Set mdl.mprPath in settings.');
			throw vscode.FileSystemError.NoPermissions('No .mpr file found');
		}

		const mdlText = new TextDecoder().decode(content);

		// Write to temp file
		const tmpFile = path.join(os.tmpdir(), `mdl-save-${Date.now()}.mdl`);
		try {
			fs.writeFileSync(tmpFile, mdlText, 'utf8');
			await this.runExec(mprFile, tmpFile);
		} finally {
			// Clean up temp file
			try { fs.unlinkSync(tmpFile); } catch {}
		}

		// Update cache with saved content
		const encoded = new TextEncoder().encode(mdlText);
		const now = Date.now();
		const cached = this.cache.get(uri.toString());
		this.cache.set(uri.toString(), { content: encoded, mtime: now, ctime: cached?.ctime ?? now });

		// Re-describe to pick up server-side normalization
		try {
			const parts = uri.path.split('/').filter(Boolean);
			if (parts.length >= 2 && uri.authority === 'describe') {
				const refreshed = await this.runDescribe(mprFile, parts[0], parts.slice(1).join('/'));
				if (!refreshed.startsWith('-- Error')) {
					const refreshedEncoded = new TextEncoder().encode(refreshed);
					// Only fire change if content actually differs
					if (mdlText !== refreshed) {
						this.cache.set(uri.toString(), { content: refreshedEncoded, mtime: Date.now(), ctime: cached?.ctime ?? now });
						this._emitter.fire([{ type: vscode.FileChangeType.Changed, uri }]);
					}
				}
			}
		} catch {
			// Refresh is best-effort; save already succeeded
		}

		vscode.window.showInformationMessage('MDL changes applied to project.');
	}

	delete(): void {
		throw vscode.FileSystemError.NoPermissions('Cannot delete');
	}

	rename(): void {
		throw vscode.FileSystemError.NoPermissions('Cannot rename');
	}

	// --- Helpers ---

	private runDescribe(mprFile: string, elementType: string, qualifiedName: string): Promise<string> {
		return new Promise<string>((resolve, reject) => {
			const args = ['describe', '-p', mprFile, elementType, qualifiedName];
			const env = { ...process.env, MXCLI_QUIET: '1' };

			cp.execFile(this.mxcliPath, args, { env, maxBuffer: 5 * 1024 * 1024 }, (err, stdout, stderr) => {
				if (err) {
					reject(new Error(stderr || err.message));
					return;
				}
				const lines = stdout.split('\n');
				const filtered = lines.filter(line => !line.startsWith('Connected to:'));
				resolve(filtered.join('\n').trimStart());
			});
		});
	}

	private runExec(mprFile: string, mdlFile: string): Promise<void> {
		return new Promise<void>((resolve, reject) => {
			const args = ['-p', mprFile, 'exec', mdlFile];
			const env = { ...process.env, MXCLI_QUIET: '1' };

			cp.execFile(this.mxcliPath, args, { env, maxBuffer: 5 * 1024 * 1024 }, (err, _stdout, stderr) => {
				if (err) {
					const msg = stderr?.trim() || err.message;
					vscode.window.showErrorMessage(`MDL save failed: ${msg}`);
					reject(new Error(msg));
					return;
				}
				resolve();
			});
		});
	}

	private async findMprFile(): Promise<string | undefined> {
		if (this.mprPath) {
			return this.mprPath;
		}
		const files = await vscode.workspace.findFiles('**/*.mpr', '**/node_modules/**', 5);
		if (files.length === 0) {
			return undefined;
		}
		return files[0].fsPath;
	}
}
