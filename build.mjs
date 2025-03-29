import * as esbuild from "esbuild";

console.log("Bundling assets...");

await esbuild.build({
	entryPoints: ["./web/app/js/app.ts", "./web/app/css/style.css"],
	bundle: true,
	minify: true,
	outdir: "./web/assets",
});
