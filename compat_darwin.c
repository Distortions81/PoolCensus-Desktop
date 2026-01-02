// Compiled only for darwin targets (via filename suffix).
//
// osxcross SDK stubs may not include __isPlatformVersionAtLeast, but Wails' Objective-C
// sources use @available() which can emit references to it. Provide a conservative
// fallback that always reports "not available" so older codepaths are used.

int __isPlatformVersionAtLeast(int platform, int major, int minor, int subminor) {
	(void)platform;
	(void)major;
	(void)minor;
	(void)subminor;
	return 0;
}

