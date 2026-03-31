"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { getToken, clearToken } from "@/lib/auth";

interface UserProfile {
  id: string;
  email: string;
  name: string;
  phone: string;
  role: string;
}

interface ProfileResponse {
  user: UserProfile;
}

export default function PerfilPage() {
  const router = useRouter();

  // Dados pessoais
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [profileLoading, setProfileLoading] = useState(true);
  const [profileSaving, setProfileSaving] = useState(false);
  const [profileSuccess, setProfileSuccess] = useState("");
  const [profileError, setProfileError] = useState("");

  // Alterar senha
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordSaving, setPasswordSaving] = useState(false);
  const [passwordSuccess, setPasswordSuccess] = useState("");
  const [passwordError, setPasswordError] = useState("");

  useEffect(() => {
    const token = getToken();
    if (!token) {
      router.push("/login");
      return;
    }

    api
      .get<ProfileResponse>("/api/profile", token)
      .then((data) => {
        setEmail(data.user.email);
        setName(data.user.name ?? "");
        setPhone(data.user.phone ?? "");
      })
      .catch((err) => {
        if (err instanceof ApiError && err.status === 401) {
          clearToken();
          router.push("/login");
        } else {
          setProfileError("Erro ao carregar perfil.");
        }
      })
      .finally(() => setProfileLoading(false));
  }, [router]);

  async function handleSaveProfile(e: React.FormEvent) {
    e.preventDefault();
    setProfileSuccess("");
    setProfileError("");

    const token = getToken();
    if (!token) {
      router.push("/login");
      return;
    }

    setProfileSaving(true);
    try {
      await api.put<ProfileResponse>("/api/profile", { name, phone }, token);
      setProfileSuccess("Dados atualizados com sucesso!");
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearToken();
        router.push("/login");
      } else {
        setProfileError("Erro ao salvar dados.");
      }
    } finally {
      setProfileSaving(false);
    }
  }

  async function handleChangePassword(e: React.FormEvent) {
    e.preventDefault();
    setPasswordSuccess("");
    setPasswordError("");

    if (newPassword.length < 8) {
      setPasswordError("A nova senha deve ter no mínimo 8 caracteres.");
      return;
    }

    if (newPassword !== confirmPassword) {
      setPasswordError("As senhas não coincidem.");
      return;
    }

    const token = getToken();
    if (!token) {
      router.push("/login");
      return;
    }

    setPasswordSaving(true);
    try {
      await api.post(
        "/api/profile/change-password",
        { current_password: currentPassword, new_password: newPassword },
        token
      );
      setPasswordSuccess("Senha alterada com sucesso!");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setPasswordError("Senha atual incorreta.");
      } else {
        setPasswordError("Erro ao alterar senha.");
      }
    } finally {
      setPasswordSaving(false);
    }
  }

  if (profileLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-gray-500">Carregando...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200 px-4 py-4">
        <div className="mx-auto max-w-lg flex items-center justify-between">
          <h1 className="text-xl font-bold text-gray-900">Meu perfil</h1>
          <button
            onClick={() => router.back()}
            className="text-sm text-gray-500 hover:text-gray-800 transition-colors"
          >
            Voltar
          </button>
        </div>
      </header>

      <main className="mx-auto max-w-lg px-4 py-8 flex flex-col gap-6">
        {/* Dados pessoais */}
        <div className="bg-white rounded-2xl border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-800 mb-4">
            Dados pessoais
          </h2>

          <form onSubmit={handleSaveProfile} className="flex flex-col gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Email
              </label>
              <input
                type="email"
                value={email}
                disabled
                className="w-full rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-500 cursor-not-allowed"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Nome
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Seu nome"
                className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Telefone
              </label>
              <input
                type="tel"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                placeholder="(11) 99999-9999"
                className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>

            {profileSuccess && (
              <p className="text-sm text-green-600">{profileSuccess}</p>
            )}
            {profileError && (
              <p className="text-sm text-red-500">{profileError}</p>
            )}

            <button
              type="submit"
              disabled={profileSaving}
              className="w-full rounded-xl py-3 text-sm font-semibold text-white transition-opacity hover:opacity-90 disabled:opacity-60 bg-primary"
            >
              {profileSaving ? "Salvando..." : "Salvar"}
            </button>
          </form>
        </div>

        {/* Alterar senha */}
        <div className="bg-white rounded-2xl border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-800 mb-4">
            Alterar senha
          </h2>

          <form onSubmit={handleChangePassword} className="flex flex-col gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Senha atual
              </label>
              <input
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                placeholder="Senha atual"
                required
                className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Nova senha
              </label>
              <input
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                placeholder="Mínimo 8 caracteres"
                required
                className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Confirmar nova senha
              </label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="Repita a nova senha"
                required
                className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>

            {passwordSuccess && (
              <p className="text-sm text-green-600">{passwordSuccess}</p>
            )}
            {passwordError && (
              <p className="text-sm text-red-500">{passwordError}</p>
            )}

            <button
              type="submit"
              disabled={passwordSaving}
              className="w-full rounded-xl py-3 text-sm font-semibold text-white transition-opacity hover:opacity-90 disabled:opacity-60 bg-accent"
            >
              {passwordSaving ? "Alterando..." : "Alterar senha"}
            </button>
          </form>
        </div>
      </main>
    </div>
  );
}
