import { useState } from "react";

// CopyText zeigt einen Wert und kopiert ihn bei Klick in die Zwischenablage.
export function CopyText({ value, className }: { value: string; className?: string }) {
  const [copied, setCopied] = useState(false);
  if (!value || value === "—") return <span className={className}>—</span>;

  const copy = async (e: React.MouseEvent) => {
    e.stopPropagation(); // nicht die Zeilenauswahl auslösen
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 1200);
    } catch { /* Clipboard nicht verfügbar */ }
  };

  return (
    <button type="button" className={`copy-text ${className ?? ""}`} onClick={copy}
      title="Zum Kopieren klicken">
      {value}
      <span className="copy-hint">{copied ? "✓ kopiert" : "⧉"}</span>
    </button>
  );
}
