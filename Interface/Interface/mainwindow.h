#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>
#include <QJsonArray>

class QNetworkAccessManager;
class QNetworkReply;
class QMediaPlayer;
class QAudioOutput;
class QListWidget;

QT_BEGIN_NAMESPACE
namespace Ui {
class MainWindow;
}
QT_END_NAMESPACE

class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

private slots:
    void cargarRolas(const QString &busqueda = "");
    void procesaRes(QNetworkReply *respuesta);
    void minarDirectorio();
    void agregaPersona();
    void consultaPersona(int id);
    void agregaGrupo();
    void cargarGrupos();
    void consultaGrupo(int id);
    void cargarIntegrantes(int idGrupo, QListWidget *lista);
    void editaRola(int fila, int columna);
    void agregaAlbum();

private:
    Ui::MainWindow *ui;
    QNetworkAccessManager *administradorRed;
    QMediaPlayer *reproductor;
    QAudioOutput *salidaAudio;
    QString      rutaActual;
    QString      rutaReproduciendo;
    int          idRolaPendiente = 0;
    void consultaRola(int id);
    QJsonArray artistasDisponibles;
    QJsonArray albumesDisponibles;
    int          idAlbumSeleccionado = 0;
};
#endif // MAINWINDOW_H
