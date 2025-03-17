import { useState, useEffect } from 'react';
import { GetSettings, UpdateSettings } from '../../wailsjs/go/app/App';

export const useSettings = () => {
  const [settings, setSettings] = useState({});
  const [darkMode, setDarkMode] = useState(true);
  const [apiError, setApiError] = useState(null);

  // Charger les paramètres
  const fetchSettings = async () => {
    try {
      const data = await GetSettings();
      console.log('Paramètres chargés:', data);
      setSettings(data);
      setApiError(null);
    } catch (error) {
      console.error('Erreur lors du chargement des paramètres:', error);
      setApiError('Impossible de charger les paramètres');
    }
  };

  // Mettre à jour les paramètres
  const handleUpdateSettings = async (newSettings) => {
    try {
      console.log('Mise à jour des paramètres:', newSettings);
      await UpdateSettings(newSettings);
      setSettings(newSettings);
      setApiError(null);
    } catch (error) {
      console.error('Erreur lors de la mise à jour des paramètres:', error);
      setApiError(`Impossible de mettre à jour les paramètres: ${error.message || error}`);
    }
  };

  // Charger les données au démarrage
  useEffect(() => {
    fetchSettings();
  }, []);

  // Gestion du darkMode via les paramètres
  useEffect(() => {
    if (settings && settings.darkMode !== undefined) {
      setDarkMode(settings.darkMode);
    }
  }, [settings]);

  return {
    settings,
    darkMode,
    apiError,
    fetchSettings,
    handleUpdateSettings,
    setDarkMode
  };
};