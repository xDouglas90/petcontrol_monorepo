import { useState, useRef, ChangeEvent } from 'react';
import { Camera, Upload, X } from 'lucide-react';
import { cn } from '@petcontrol/ui';

interface ImageUploadProps {
  value?: string; // URL da imagem atual (persistida)
  onChange?: (objectKey: string) => void; // Mantido para compatibilidade se necessário
  onFileSelect?: (file: File | null) => void; // Novo: notifica o pai sobre o arquivo selecionado
  module: 'pets' | 'companies' | 'identifications';
  className?: string;
  label?: string;
}

export function ImageUpload({
  value,
  onChange,
  onFileSelect,
  className,
  label,
}: ImageUploadProps) {
  const [localPreview, setLocalPreview] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const preview = localPreview || value || null;

  function handleFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;

    // Criar preview local
    const localUrl = URL.createObjectURL(file);
    setLocalPreview(localUrl);

    // Notificar o pai
    if (onFileSelect) {
      onFileSelect(file);
    }
  }

  function removeImage() {
    setLocalPreview(null);
    if (onFileSelect) {
      onFileSelect(null);
    }
    if (onChange) {
      onChange('');
    }
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  }

  return (
    <div className={cn('space-y-2', className)}>
      {label && (
        <span className="text-sm font-medium text-slate-200">{label}</span>
      )}

      <div
        className={cn(
          'relative group flex flex-col items-center justify-center aspect-square w-32 rounded-3xl border-2 border-dashed transition-all overflow-hidden bg-white/5',
          preview
            ? 'border-primary/40'
            : 'border-white/10 hover:border-primary/30 hover:bg-white/10 cursor-pointer',
        )}
        onClick={() => !preview && fileInputRef.current?.click()}
      >
        {preview ? (
          <>
            <img
              src={preview}
              alt="Preview"
              className="w-full h-full object-cover transition-transform group-hover:scale-110"
            />
            <button
              type="button"
              title="Remover imagem"
              onClick={(e) => {
                e.stopPropagation();
                removeImage();
              }}
              className="absolute top-2 right-2 p-1.5 rounded-xl bg-rose-500/80 text-white opacity-0 group-hover:opacity-100 transition-opacity hover:bg-rose-600"
            >
              <X size={14} />
            </button>
            <div className="absolute inset-0 bg-black/40 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none">
              <button
                type="button"
                title="Trocar imagem"
                className="p-2 rounded-full bg-primary text-slate-950 scale-75 group-hover:scale-100 transition-transform pointer-events-auto"
                onClick={(e) => {
                  e.stopPropagation();
                  fileInputRef.current?.click();
                }}
              >
                <Camera size={20} />
              </button>
            </div>
          </>
        ) : (
          <div className="flex flex-col items-center gap-2 text-slate-400 p-4 text-center">
            <Upload size={24} />
            <span className="text-[10px] font-medium uppercase tracking-wider">
              Selecionar Foto
            </span>
          </div>
        )}
      </div>

      <input
        type="file"
        ref={fileInputRef}
        className="hidden"
        accept="image/*"
        title="Selecionar imagem"
        onChange={handleFileChange}
      />
    </div>
  );
}
