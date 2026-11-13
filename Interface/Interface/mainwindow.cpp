#include "mainwindow.h"
#include "./ui_mainwindow.h"
#include <QNetworkAccessManager>
#include <QNetworkRequest>
#include <QNetworkReply>
#include <QUrl>
#include <QJsonDocument>
#include <QJsonArray>
#include <QJsonObject>
#include <QDebug>
#include <QTableWidgetItem>
#include <QUrlQuery>
#include <QFileDialog>
#include <QMessageBox>
#include <QListWidgetItem>
#include <QSet>
#include <QPixmap>
#include <QIcon>
#include <QMediaPlayer>
#include <QAudioOutput>
#include <QPushButton>
#include <QSlider>
#include <QDialog>
#include <QFormLayout>
#include <QLineEdit>
#include <QDialogButtonBox>
#include <QVBoxLayout>
#include <QDate>
#include <QDateEdit>
#include <QCheckBox>
#include <QHBoxLayout>
#include <QMessageBox>

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , ui(new Ui::MainWindow)
    , administradorRed(new QNetworkAccessManager(this))
    , reproductor(new QMediaPlayer(this))
    , salidaAudio(new QAudioOutput(this))
{
    ui->setupUi(this);
    reproductor->setAudioOutput(salidaAudio);
    connect(administradorRed, &QNetworkAccessManager::finished, this, &MainWindow::procesaRes);
    cargarRolas();
    connect(ui->songsButton, &QPushButton::clicked, this, [this]() {
        ui->pagesStackedWidget->setCurrentWidget(ui->songsPage);
        cargarRolas(ui->searchLineEdit->text());
    });
    connect(ui->searchLineEdit, &QLineEdit::returnPressed, this, [this]() {
        cargarRolas(ui->searchLineEdit->text());
    });
    connect(ui->foldersButton, &QPushButton::clicked, this, &MainWindow::minarDirectorio);
    connect(ui->albumsButton, &QPushButton::clicked, this, [this]() {
        ui->pagesStackedWidget->setCurrentWidget(ui->albumsPage);
    });
    connect(ui->songsTable, &QTableWidget::currentCellChanged, this, [this](int fila, int, int, int) {
                if (fila < 0) {
                    return;
                }
                QTableWidgetItem *tituloItem = ui->songsTable->item(fila, 0);
                QTableWidgetItem *artistaItem = ui->songsTable->item(fila, 1);

                if (!tituloItem || !artistaItem) {
                    return;
                }
                QString titulo = ui->songsTable->item(fila, 0)->text();
                QString artista = ui->songsTable->item(fila, 1)->text();
                rutaActual = tituloItem->data(Qt::UserRole).toString();
                ui->currentTrackLabel->setText(titulo + " - " + artista);
    });
    connect(ui->playPauseButton, &QPushButton::clicked,
            this, [this]() {

                if (rutaActual.isEmpty()) {
                    return;
                }
                if (rutaActual != rutaReproduciendo) {
                    rutaReproduciendo = rutaActual;
                    reproductor->setSource(QUrl::fromLocalFile(rutaActual));
                    reproductor->play();
                    return;
                }
                if (reproductor->playbackState() == QMediaPlayer::PlayingState) {
                    reproductor->pause();
                } else {
                    reproductor->play();
                }
    });
    auto cambiarRola = [this](int desplazamiento) {
        int filaActual = ui->songsTable->currentRow();
        int nuevaFila = filaActual + desplazamiento;
        if (nuevaFila < 0 || nuevaFila >= ui->songsTable->rowCount()) {
            return;
        }

        ui->songsTable->setCurrentCell(nuevaFila, 0);
        QTableWidgetItem *tituloItem = ui->songsTable->item(nuevaFila, 0);
        if (!tituloItem) {
            return;
        }

        rutaActual = tituloItem->data(Qt::UserRole).toString();
        if (rutaActual.isEmpty()) {
            return;
        }

        rutaReproduciendo = rutaActual;
        reproductor->setSource(QUrl::fromLocalFile(rutaActual));
        reproductor->play();
    };
    connect(ui->previousButton, &QPushButton::clicked,
            this, [cambiarRola]() {
                cambiarRola(-1);
    });

    connect(ui->nextButton, &QPushButton::clicked,
            this, [cambiarRola]() {
                cambiarRola(1);
    });
    connect(ui->volumeSlider, &QSlider::valueChanged, this, [this](int volumen) {
            salidaAudio->setVolume(volumen / 100.0);
    });
    connect(reproductor, &QMediaPlayer::durationChanged, this, [this](qint64 duracion) {
        ui->progressSlider->setRange(0, duracion);
    });
    connect(reproductor, &QMediaPlayer::positionChanged, this, [this](qint64 posicion) {
        ui->progressSlider->setValue(posicion);
    });
    connect(ui->progressSlider, &QSlider::sliderMoved, this, [this](int posicion) {
        reproductor->setPosition(posicion);
    });
    connect(reproductor, &QMediaPlayer::mediaStatusChanged, this, [this](QMediaPlayer::MediaStatus estado) {
        if (estado == QMediaPlayer::EndOfMedia) {
            ui->progressSlider->setValue(0);
        }
    });
    connect(ui->artistsButton, &QPushButton::clicked, this, [this]() {
        ui->pagesStackedWidget->setCurrentWidget(ui->artistsPage);
    });
    connect(ui->artistsButton, &QPushButton::clicked, this, [this]() {
        ui->pagesStackedWidget->setCurrentWidget(ui->artistsPage);
        QUrl direccion("http://127.0.0.1:8080/api/artistas");
        QNetworkRequest peticion(direccion);
        administradorRed->get(peticion);
    });
    connect(ui->addArtistButton, &QPushButton::clicked, this, &MainWindow::agregaPersona);
}

MainWindow::~MainWindow()
{
    delete ui;
}

void MainWindow::cargarRolas(const QString &busqueda)
{
    QUrl
        direccion("http://127.0.0.1:8080/api/rolas");
    QUrlQuery parametros;
    parametros.addQueryItem("q", busqueda);
    direccion.setQuery(parametros);
    QNetworkRequest peticion(direccion);
    administradorRed->get(peticion);
}

void MainWindow::procesaRes(QNetworkReply *respuesta)
{
    QByteArray datos = respuesta->readAll();

    if (respuesta->url().path() == "/api/minar") {
        if (respuesta->error() != QNetworkReply::NoError) {
            QMessageBox::critical(
                this,
                "Ocurrió un error al minar",
                QString::fromUtf8(datos)
                );
        } else {
            QJsonDocument documento = QJsonDocument::fromJson(datos);

            if (documento.isObject()) {
                QJsonObject resultado = documento.object();

                QString estado = resultado["status"].toString();
                QString mensaje = resultado["message"].toString();

                if (estado == "ok") {
                    QMessageBox::information(
                        this,
                        "Minado completado",
                        mensaje
                        );

                    cargarRolas(ui->searchLineEdit->text());
                } else {
                    QMessageBox::warning(
                        this,
                        "Resultado del minado",
                        mensaje
                        );
                }
            }
        }

        respuesta->deleteLater();
        return;
    }

    if (respuesta->url().path() == "/api/artistas") {
        QJsonDocument documento = QJsonDocument::fromJson(datos);
        if (!documento.isArray()) {
            respuesta->deleteLater();
            return;
        }

        QJsonArray artistas = documento.array();

        ui->artistsTable->setRowCount(0);

        for (const QJsonValue &valor : artistas) {
            QJsonObject artista = valor.toObject();

            int fila = ui->artistsTable->rowCount();
            ui->artistsTable->insertRow(fila);

            int tipo = artista["Tipo"].toInt();

            QString nombreTipo;
            if (tipo == 0) {
                nombreTipo = "Persona";
            } else if (tipo == 1) {
                nombreTipo = "Grupo";
            } else {
                nombreTipo = "Desconocido";
            }

            QTableWidgetItem *tipoItem = new QTableWidgetItem(nombreTipo);
            QTableWidgetItem *nombreItem = new QTableWidgetItem(
                artista["Nombre"].toString()
                );

            nombreItem->setData(Qt::UserRole, artista["Id"].toInt());
            ui->artistsTable->setItem(fila, 0, tipoItem);
            ui->artistsTable->setItem(fila, 1, nombreItem);
        }

        respuesta->deleteLater();
        return;
    }
    if (respuesta->url().path().startsWith("/api/portadas/")) {
        if (respuesta->error() == QNetworkReply::NoError) {

            int idAlbum = respuesta->url().path()
            .section('/', -1).toInt();

            QPixmap portada;
            portada.loadFromData(datos);

            for (int i = 0; i < ui->albumsListWidget->count(); i++) {
                QListWidgetItem *elemento = ui->albumsListWidget->item(i);

                if (elemento->data(Qt::UserRole).toInt() == idAlbum) {
                    elemento->setIcon(QIcon(portada));
                    break;
                }
            }
        }

        respuesta->deleteLater();
        return;
    }

    if (respuesta->error() != QNetworkReply::NoError) {
        respuesta->deleteLater();
        return;
    }

    QJsonDocument documento = QJsonDocument::fromJson(datos);

    if (!documento.isArray()) {
        respuesta->deleteLater();
        return;
    }

    QJsonArray rolas = documento.array();

    QUrlQuery parametros(respuesta->url());
    if (parametros.queryItemValue("q").isEmpty()) {
        ui->albumsListWidget->clear();

        QSet<int> albumesAgregados;

        for (const QJsonValue &valor : rolas) {
            QJsonObject rola = valor.toObject();

            int idAlbum = rola["album_id"].toInt();
            if (idAlbum <= 0 || albumesAgregados.contains(idAlbum)) {
                continue;
            }

            albumesAgregados.insert(idAlbum);

            QString nombre = rola["album"].toString();
            QString artista = rola["artist"].toString();
            QString texto = nombre;
            if (nombre != artista) {
                texto += "\n" + artista;
            }

            QListWidgetItem *elemento = new QListWidgetItem(texto);
            elemento->setData(Qt::UserRole, idAlbum);
            elemento->setTextAlignment(Qt::AlignHCenter);
            ui->albumsListWidget->addItem(elemento);
            QUrl direccionPortada("http://127.0.0.1:8080/api/portadas/" + QString::number(idAlbum));
            QNetworkRequest peticionPortada(direccionPortada);
            administradorRed->get(peticionPortada);
        }
    }

    ui->songsTable->setRowCount(0);

    for (const QJsonValue &valor : rolas) {
        QJsonObject rola = valor.toObject();

        int fila = ui->songsTable->rowCount();

        ui->songsTable->insertRow(fila);
        QTableWidgetItem *tituloItem = new QTableWidgetItem(rola["title"].toString());
        tituloItem->setData(Qt::UserRole, rola["path"].toString());
        ui->songsTable->setItem(fila, 0, tituloItem);

        ui->songsTable->setItem(fila, 1, new QTableWidgetItem(rola["artist"].toString()));
        ui->songsTable->setItem(fila, 2, new QTableWidgetItem(rola["album"].toString()));
        ui->songsTable->setItem(fila, 3, new QTableWidgetItem(QString::number(rola["year"].toInt())));
        ui->songsTable->setItem(fila, 4, new QTableWidgetItem(rola["genre"].toString()));
    }

    respuesta->deleteLater();
}

void MainWindow::minarDirectorio()
{
    QString directorio = QFileDialog::getExistingDirectory(
        this,
        "Seleccionar carpeta de rolas"
        );

    if (directorio.isEmpty()) {
        return;
    }

    QUrl direccion("http://127.0.0.1:8080/api/minar");
    QNetworkRequest peticion(direccion);

    peticion.setHeader(
        QNetworkRequest::ContentTypeHeader,
        "application/json"
        );

    QJsonObject solicitud;
    solicitud["directorio"] = directorio;

    QJsonDocument documento(solicitud);
    QByteArray datos = documento.toJson();

    administradorRed->post(peticion, datos);
}

void MainWindow::agregaPersona()
{
    QDialog dialogo(this);
    dialogo.setWindowTitle("Agregar artista");
    QLineEdit *nombreArtistico = new QLineEdit(&dialogo);
    QLineEdit *nombreReal = new QLineEdit(&dialogo);


    QCheckBox *tieneNacimiento = new QCheckBox("Especificar fecha");
    QCheckBox *tieneFallecimiento = new QCheckBox("Especificar fecha");
    tieneFallecimiento->setEnabled(false);
    QDateEdit *fechaNacimiento = new QDateEdit(&dialogo);
    QDateEdit *fechaFallecimiento = new QDateEdit(&dialogo);


    fechaNacimiento->setDisplayFormat("dd/MM/yyyy");
    fechaFallecimiento->setDisplayFormat("dd/MM/yyyy");
    fechaNacimiento->setCalendarPopup(true);
    fechaFallecimiento->setCalendarPopup(true);
    fechaNacimiento->setEnabled(false);
    fechaFallecimiento->setEnabled(false);

    connect(tieneNacimiento, &QCheckBox::toggled, fechaNacimiento, &QDateEdit::setEnabled);
    connect(tieneFallecimiento, &QCheckBox::toggled, fechaFallecimiento, &QDateEdit::setEnabled);
    connect(tieneNacimiento, &QCheckBox::toggled, tieneFallecimiento, &QCheckBox::setEnabled);
    connect(tieneNacimiento, &QCheckBox::toggled, this, [tieneFallecimiento](bool marcado) {
                if (!marcado)
                    tieneFallecimiento->setChecked(false);
    });

    QFormLayout *formulario = new QFormLayout;

    formulario->addRow("Nombre artístico:", nombreArtistico);
    formulario->addRow("Nombre real:", nombreReal);
    QHBoxLayout *filaNacimiento = new QHBoxLayout;
    filaNacimiento->addWidget(tieneNacimiento);
    filaNacimiento->addWidget(fechaNacimiento);

    QHBoxLayout *filaFallecimiento = new QHBoxLayout;
    filaFallecimiento->addWidget(tieneFallecimiento);
    filaFallecimiento->addWidget(fechaFallecimiento);

    formulario->addRow("Fecha de nacimiento:", filaNacimiento);
    formulario->addRow("Fecha de fallecimiento:", filaFallecimiento);

    QDialogButtonBox *botones = new QDialogButtonBox(
        QDialogButtonBox::Ok | QDialogButtonBox::Cancel
        );
    connect(botones, &QDialogButtonBox::accepted,
            &dialogo, &QDialog::accept);
    connect(botones, &QDialogButtonBox::rejected,
            &dialogo, &QDialog::reject);

    QVBoxLayout *layout = new QVBoxLayout(&dialogo);
    layout->addLayout(formulario);
    layout->addWidget(botones);

    if(dialogo.exec() != QDialog::Accepted)
        return;

    if (tieneNacimiento->isChecked() &&
        tieneFallecimiento->isChecked()) {

        if (fechaNacimiento->date() >= fechaFallecimiento->date()) {
            QMessageBox::warning(this,
                "Fechas inválidas", "La fecha de fallecimiento debe ser posterior a la fecha de nacimiento."
                );

            return;
        }
    }
}
