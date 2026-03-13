const PALETTE = [
	'#6366f1', // indigo
	'#8b5cf6', // violet
	'#ec4899', // pink
	'#f59e0b', // amber
	'#10b981', // emerald
	'#3b82f6', // blue
	'#ef4444', // red
	'#14b8a6', // teal
];

export function orgAvatarColor(slug: string): string {
	let hash = 0;
	for (const char of slug) hash = (hash * 31 + char.charCodeAt(0)) & 0xffffffff;
	return PALETTE[Math.abs(hash) % PALETTE.length];
}

export function orgInitials(name: string): string {
	return name
		.split(/\s+/)
		.slice(0, 2)
		.map((w) => w[0])
		.join('')
		.toUpperCase();
}

export function slugify(name: string): string {
	return name
		.toLowerCase()
		.replace(/[^a-z0-9\s-]/g, '')
		.trim()
		.replace(/\s+/g, '-')
		.replace(/-+/g, '-');
}
