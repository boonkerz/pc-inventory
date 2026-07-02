import { ReactNode, useEffect } from "react";

// Modal zeigt Inhalt in einem großen Overlay (~80% des Bildschirms). Schließt
// per Escape, Klick auf den Hintergrund oder den Schließen-Button.
export function Modal({ onClose, children }: { onClose: () => void; children: ReactNode }) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === "Escape") onClose(); };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal-panel" onClick={(e) => e.stopPropagation()}>
        <button className="modal-close" onClick={onClose} aria-label="Schließen">✕</button>
        <div className="modal-body">{children}</div>
      </div>
    </div>
  );
}
