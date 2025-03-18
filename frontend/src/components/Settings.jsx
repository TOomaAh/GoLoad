import React, { useState, useEffect } from 'react';
import { Settings, ArrowLeft } from 'lucide-react';

// Fonction pour détecter si l'application tourne en mode desktop (avec Wails)
const isDesktopApp = () => {
  return typeof window !== 'undefined' && window.runtime !== undefined;
};

const SettingsPage = ({ settings, darkMode, onUpdateSettings, onBackClick }) => {
  // State local pour suivre le mode sombre dans ce composant
  const [localDarkMode, setLocalDarkMode] = useState(darkMode);
  
  // Mettre à jour le state local lorsque les props changent
  useEffect(() => {
    setLocalDarkMode(darkMode);
  }, [darkMode]);

  const handleToggle = (key) => {
    const newSettings = { ...settings, [key]: !settings[key] };
    
    // Si c'est le mode sombre qui est mis à jour, mettre à jour le state local également
    if (key === 'darkMode') {
      setLocalDarkMode(!localDarkMode);
    }
    
    onUpdateSettings(newSettings);
  };

  const handleInputChange = (key, value) => {
    const newSettings = { ...settings, [key]: value };
    onUpdateSettings(newSettings);
  };

  // Fonction pour gérer le clic sur le bouton Parcourir
  const handleBrowse = async () => {
    try {
      if (isDesktopApp()) {
        // Version desktop: utiliser l'API Wails pour ouvrir un sélecteur de dossier natif
        // Assumons qu'il y a une fonction dans l'API Wails appelée OpenDirectoryDialog
        const selectedDir = await window.runtime.OpenDirectoryDialog();
        if (selectedDir) {
          handleInputChange('downloadPath', selectedDir);
        }
      } else {
        // Version web: utiliser l'API web standard
        // Créer un élément input de type directory (note: support limité dans les navigateurs)
        const input = document.createElement('input');
        input.type = 'file';
        input.webkitdirectory = true; // Non-standard, fonctionne dans Chrome, Edge, Safari
        input.directory = true; // Firefox

        // Gérer la sélection
        input.onchange = (e) => {
          if (e.target.files.length > 0) {
            // Récupérer le chemin du dossier sélectionné
            // Ceci obtiendra juste le nom du dossier, pas le chemin complet
            // (restrictions de sécurité des navigateurs)
            const path = e.target.files[0].webkitRelativePath.split('/')[0];
            handleInputChange('downloadPath', path);
          }
        };

        // Déclencher le dialogue de sélection
        input.click();
      }
    } catch (error) {
      console.error('Erreur lors de la sélection du dossier:', error);
      // Vous pourriez vouloir montrer une notification d'erreur à l'utilisateur ici
    }
  };

  return (
    <div className={`flex flex-col h-screen ${localDarkMode ? 'bg-gray-900 text-gray-100' : 'bg-gray-50 text-gray-800'}`}>
      {/* Header pour les paramètres */}
      <header className={`${localDarkMode ? 'bg-gray-800' : 'bg-white'} shadow-md px-6 py-4`}>
        <div className="flex items-center">
          <button
            onClick={onBackClick}
            className={`p-2 mr-4 rounded-md ${localDarkMode ? 'hover:bg-gray-700' : 'hover:bg-gray-100'}`}
            title="Retour"
          >
            <ArrowLeft size={22} />
          </button>
          <h1 className="text-2xl font-bold flex items-center">
            <Settings className="mr-3" size={28} />
            Paramètres
          </h1>
        </div>
      </header>

      {/* Contenu des paramètres */}
      <div className="flex-1 overflow-auto p-6">
        <div className="max-w-3xl mx-auto space-y-8">
          {/* Section apparence */}
          <SettingsSection title="Apparence" darkMode={localDarkMode}>
            <ToggleSetting
              id="darkMode"
              label="Mode sombre"
              value={localDarkMode}
              onChange={() => handleToggle('darkMode')}
            />
          </SettingsSection>
          
          {/* Section téléchargements */}
          <SettingsSection title="Téléchargements" darkMode={localDarkMode}>
            <div className="flex flex-col">
              <label htmlFor="downloadPath" className="font-medium mb-2">
                Dossier de téléchargement par défaut
              </label>
              <div className="flex">
                <input
                  type="text"
                  id="downloadPath"
                  className={`flex-1 px-3 py-2 border ${
                    localDarkMode 
                      ? 'border-gray-600 bg-gray-700 focus:border-blue-500' 
                      : 'border-gray-300 focus:border-blue-500'
                  } rounded-l-md focus:outline-none focus:ring-1 focus:ring-blue-500`}
                  value={settings.downloadPath || ''}
                  onChange={(e) => handleInputChange('downloadPath', e.target.value)}
                />
                <button
                  onClick={handleBrowse}
                  className={`px-4 py-2 ${
                    localDarkMode ? 'bg-gray-700 hover:bg-gray-600' : 'bg-gray-200 hover:bg-gray-300'
                  } rounded-r-md`}
                  title="Parcourir"
                >
                  Parcourir
                </button>
              </div>
            </div>
            
            <div className="flex items-center justify-between">
              <label htmlFor="maxConcurrentDownloads" className="font-medium">
                Téléchargements simultanés maximum
              </label>
              <div className="w-20">
                <input
                  type="number"
                  id="maxConcurrentDownloads"
                  min="1"
                  max="10"
                  className={`w-full px-3 py-2 border ${
                    localDarkMode 
                      ? 'border-gray-600 bg-gray-700 focus:border-blue-500' 
                      : 'border-gray-300 focus:border-blue-500'
                  } rounded-md focus:outline-none focus:ring-1 focus:ring-blue-500`}
                  value={settings.maxConcurrentDownloads || 3}
                  onChange={(e) => handleInputChange('maxConcurrentDownloads', parseInt(e.target.value, 10))}
                />
              </div>
            </div>
            <ToggleSetting
              id="autoStartDownloads"
              label="Démarrer automatiquement les téléchargements"
              value={settings.autoStartDownload || false}
              onChange={() => handleToggle('autoStartDownloads')}
            />
            <ToggleSetting
              id="autoRetry"
              label="Réessayer automatiquement après une erreur"
              value={settings.autoRetry || false}
              onChange={() => handleToggle('autoRetry')}
            />
          </SettingsSection>
          
          {/* Section notifications */}
          <SettingsSection title="Notifications" darkMode={localDarkMode}>
            <ToggleSetting
              id="notifyOnComplete"
              label="Notifier à la fin des téléchargements"
              value={settings.notifyOnComplete || false}
              onChange={() => handleToggle('notifyOnComplete')}
            />
            
            <ToggleSetting
              id="notifyOnError"
              label="Notifier en cas d'erreur"
              value={settings.notifyOnError || false}
              onChange={() => handleToggle('notifyOnError')}
            />
          </SettingsSection>
        </div>
      </div>
    </div>
  );
};

// Composant pour une section de paramètres
const SettingsSection = ({ title, children, darkMode }) => {
  return (
    <div className={`${darkMode ? 'bg-gray-800' : 'bg-white'} rounded-lg p-6 shadow-sm`}>
      <h2 className="text-xl font-semibold mb-4">{title}</h2>
      <div className="space-y-4">
        {children}
      </div>
    </div>
  );
};

// Composant pour un paramètre de type interrupteur (toggle)
const ToggleSetting = ({ id, label, value, onChange }) => {
  return (
    <div className="flex items-center justify-between">
      <label htmlFor={id} className="font-medium cursor-pointer">
        {label}
      </label>
      <div 
        className="relative inline-block w-12 align-middle select-none cursor-pointer"
        onClick={onChange}
      >
        <input
          type="checkbox"
          id={id}
          checked={value}
          onChange={onChange}
          className="hidden"
        />
        <div 
          className={`block w-12 h-6 rounded-full ${value ? 'bg-blue-500' : 'bg-gray-300'}`}
        ></div>
        <div 
          className={`absolute left-1 top-1 bg-white w-4 h-4 rounded-full transition-transform transform ${
            value ? 'translate-x-6' : ''
          }`}
        ></div>
      </div>
    </div>
  );
};

export default SettingsPage;