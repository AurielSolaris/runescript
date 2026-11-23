interface CodeCopy {
  copied: boolean;
  copy(this: CodeCopy & { $root: HTMLElement }): Promise<void>;
}

export function registerCodeCopy(alpine: { data(name: string, component: () => CodeCopy): void }) {
  alpine.data("codeCopy", () => ({
    copied: false,
    async copy() {
      const code = this.$root.querySelector("pre code");
      if (!code) return;
      try {
        await navigator.clipboard.writeText(code.textContent ?? "");
        this.copied = true;
      } catch {
        this.copied = false;
      }
    },
  }));
}
