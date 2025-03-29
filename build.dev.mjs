import * as esbuild from "esbuild";

let ctx = await esbuild.context({
	entryPoints: ["./web/app/js/app.ts", "./web/app/css/style.css"],
	bundle: true,
	sourcemap: true,
	outdir: "./web/assets",
});

await ctx.watch();

console.log("Watching assets...");
