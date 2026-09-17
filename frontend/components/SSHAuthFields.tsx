"use client";

import { useI18n } from "@/lib/i18n/context";

export type SSHAuthMode = "password" | "private_key";

export type SSHAuthValues = {
  mode: SSHAuthMode;
  password: string;
  privateKey: string;
  passphrase: string;
};

export function sshAuthBody(values: SSHAuthValues): {
  password?: string;
  private_key?: string;
  private_key_passphrase?: string;
} {
  if (values.mode === "private_key") {
    const body: {
      private_key?: string;
      private_key_passphrase?: string;
    } = { private_key: values.privateKey };
    if (values.passphrase) {
      body.private_key_passphrase = values.passphrase;
    }
    return body;
  }
  return { password: values.password };
}

export function sshAuthReady(values: SSHAuthValues): boolean {
  if (values.mode === "private_key") {
    return values.privateKey.trim().length > 0;
  }
  return values.password.length > 0;
}

type Props = {
  idPrefix: string;
  values: SSHAuthValues;
  onChange: (next: SSHAuthValues) => void;
  disabled?: boolean;
};

export function SSHAuthFields({ idPrefix, values, onChange, disabled }: Props) {
  const { t } = useI18n();

  return (
    <>
      <div className="field">
        <span className="label">{t("servers.sshAuthMethod")}</span>
        <div className="server-status-toggles" style={{ marginTop: "0.35rem" }}>
          <label>
            <input
              type="radio"
              name={`${idPrefix}-auth-mode`}
              checked={values.mode === "password"}
              disabled={disabled}
              onChange={() =>
                onChange({
                  ...values,
                  mode: "password",
                  privateKey: "",
                  passphrase: "",
                })
              }
            />{" "}
            {t("servers.sshAuthPassword")}
          </label>
          <label>
            <input
              type="radio"
              name={`${idPrefix}-auth-mode`}
              checked={values.mode === "private_key"}
              disabled={disabled}
              onChange={() =>
                onChange({
                  ...values,
                  mode: "private_key",
                  password: "",
                })
              }
            />{" "}
            {t("servers.sshAuthPrivateKey")}
          </label>
        </div>
      </div>

      {values.mode === "password" ? (
        <div className="field">
          <label className="label" htmlFor={`${idPrefix}-pass`}>
            {t("servers.sshPassword")}
          </label>
          <input
            id={`${idPrefix}-pass`}
            className="input"
            type="password"
            autoComplete="new-password"
            value={values.password}
            disabled={disabled}
            onChange={(e) => onChange({ ...values, password: e.target.value })}
            required
          />
        </div>
      ) : (
        <>
          <div className="field">
            <label className="label" htmlFor={`${idPrefix}-key`}>
              {t("servers.sshPrivateKey")}
            </label>
            <textarea
              id={`${idPrefix}-key`}
              className="input"
              rows={6}
              spellCheck={false}
              autoComplete="off"
              placeholder={t("servers.sshPrivateKeyPlaceholder")}
              value={values.privateKey}
              disabled={disabled}
              onChange={(e) => onChange({ ...values, privateKey: e.target.value })}
              required
              style={{ fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace", fontSize: "0.8rem" }}
            />
            <p className="muted" style={{ margin: "0.35rem 0 0", fontSize: "0.85rem" }}>
              {t("servers.sshPrivateKeyHint")}
            </p>
          </div>
          <div className="field">
            <label className="label" htmlFor={`${idPrefix}-passphrase`}>
              {t("servers.sshPrivateKeyPassphrase")}
            </label>
            <input
              id={`${idPrefix}-passphrase`}
              className="input"
              type="password"
              autoComplete="off"
              value={values.passphrase}
              disabled={disabled}
              onChange={(e) => onChange({ ...values, passphrase: e.target.value })}
            />
            <p className="muted" style={{ margin: "0.35rem 0 0", fontSize: "0.85rem" }}>
              {t("servers.sshPrivateKeyPassphraseHint")}
            </p>
          </div>
        </>
      )}
    </>
  );
}
