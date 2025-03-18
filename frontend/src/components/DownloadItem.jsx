import React from "react";
import { Pause, Play, X, Trash2 } from "lucide-react";
import {
  getStatusLabel,
  getStatusClasses,
  getProgressBarColor,
} from "../utils/formatters";

const DownloadItem = ({
  download,
  darkMode,
  onDownloadAction,
  onRemoveDownload,
}) => {
  return (
    <div
      className={`${
        darkMode ? "bg-gray-800" : "bg-white"
      } rounded-lg shadow-sm p-4 transition-all hover:shadow-md`}
    >
      <div className="flex justify-between items-start mb-3">
        <div className="flex-1 pr-4">
          <h3 className="font-medium text-lg truncate">{download.filename}</h3>

          <div
            className={`flex text-sm ${
              darkMode ? "text-gray-400" : "text-gray-500"
            } mt-1`}
          >
            <span>
              {download.downloaded_readable || `${download.downloaded} B`} /{" "}
              {download.total_size_readable || `${download.size} B`}
            </span>
            <span className="mx-2">•</span>
            <span>{download.speed_readable || `${download.speed} B/s`}</span>
            {download.status === "downloading" && (
              <>
                <span className="mx-2">•</span>
                <span>Reste: {download.eta || "--:--"}</span>
              </>
            )}
          </div>
        </div>

        <div className="flex space-x-1">
          {download.status === "downloading" && (
            <button
              onClick={() => onDownloadAction(download.ID, "pause")}
              className={`p-2 rounded-md ${
                darkMode ? "hover:bg-gray-700" : "hover:bg-gray-100"
              }`}
              title="Mettre en pause"
            >
              <Pause size={18} />
            </button>
          )}
          {download.status === "paused" ||
            (download.status === "pending" && (
              <button
                onClick={() => onDownloadAction(download.ID, "resume")}
                className={`p-2 rounded-md ${
                  darkMode ? "hover:bg-gray-700" : "hover:bg-gray-100"
                }`}
                title="Reprendre"
              >
                <Play size={18} />
              </button>
            ))}
          {download.status !== "completed" && download.status !== "error" && download.status !== "cancelled" && (
            <button
              onClick={() => onDownloadAction(download.ID, "cancel")}
              className={`p-2 rounded-md ${
                darkMode ? "hover:bg-gray-700" : "hover:bg-gray-100"
              }`}
              title="Annuler"
            >
              <X size={18} />
            </button>
          )}
          <button
            onClick={() => onRemoveDownload(download.ID)}
            className={`p-2 rounded-md ${
              darkMode ? "hover:bg-gray-700" : "hover:bg-gray-100"
            }`}
            title="Supprimer"
          >
            <Trash2 size={18} />
          </button>
        </div>
      </div>

      <div
        className={`${
          darkMode ? "bg-gray-700" : "bg-gray-200"
        } rounded-full h-2 mt-1`}
      >
        <div
          className={`h-2 rounded-full transition-all ${getProgressBarColor(
            download.status
          )}`}
          style={{ width: `${download.progress}%` }}
        ></div>
      </div>

      <div className="mt-2">
        <span
          className={`px-2 py-1 text-xs rounded-full ${getStatusClasses(
            download.status,
            darkMode
          )}`}
        >
          {getStatusLabel(download.status)}
        </span>
        {download.error_message && (
          <span
            className={`ml-2 text-xs ${
              darkMode ? "text-red-400" : "text-red-600"
            }`}
          >
            {download.error_message}
          </span>
        )}
      </div>
    </div>
  );
};

export default DownloadItem;
