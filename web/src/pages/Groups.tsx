import { FormEvent, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "../api";
import { useI18n } from "../i18n";
import type { Group } from "../types";
import { useAuth } from "../auth";

export function Groups() {
  const { t } = useI18n();
  const qc = useQueryClient();
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  const { data: groups } = useQuery({ queryKey: ["groups"], queryFn: () => api.get<Group[]>("/groups") });

  const create = useMutation({
    mutationFn: () => api.post<Group>("/groups", { name, description }),
    onSuccess: () => {
      setName("");
      setDescription("");
      qc.invalidateQueries({ queryKey: ["groups"] });
    },
  });
  const remove = useMutation({
    mutationFn: (id: string) => api.del(`/groups/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["groups"] }),
  });

  const submit = (e: FormEvent) => {
    e.preventDefault();
    if (name.trim()) create.mutate();
  };

  return (
    <div className="page">
      <header className="page-head">
        <div>
          <h1>{t("Tags")}</h1>
          <p className="muted">{t("Freie Labels zum Querschneiden (n:m) – unabhängig von Client/Standort.")}</p>
        </div>
      </header>

      {isAdmin && (
        <form className="card inline-form" onSubmit={submit}>
          <input placeholder={t("Tag-Name")} value={name} onChange={(e) => setName(e.target.value)} />
          <input placeholder={t("Beschreibung (optional)")} value={description} onChange={(e) => setDescription(e.target.value)} />
          <button className="btn primary" type="submit" disabled={create.isPending}>{t("Anlegen")}</button>
        </form>
      )}

      <div className="cards">
        {(groups ?? []).map((g) => (
          <div className="card group-card" key={g.id}>
            <div>
              <div className="group-name">{g.name}</div>
              <div className="muted">{g.description || "—"}</div>
            </div>
            <div className="group-meta">
              <span className="count">{g.device_count ?? 0} Geräte</span>
              {isAdmin && (
                <button className="btn ghost sm" onClick={() => confirm(`Tag „${g.name}" löschen?`) && remove.mutate(g.id)}>
                  Löschen
                </button>
              )}
            </div>
          </div>
        ))}
        {(groups ?? []).length === 0 && <p className="muted">{t("Noch keine Tags.")}</p>}
      </div>
    </div>
  );
}
