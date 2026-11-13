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
#include <QLabel>

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
        QUrl direccion("http://127.0.0.1:8080/api/artistas");
        QNetworkRequest peticion(direccion);
        administradorRed->get(peticion);
    });
    connect(ui->addArtistButton, &QPushButton::clicked, this, &MainWindow::agregaPersona);
    connect(ui->artistsTable, &QTableWidget::cellDoubleClicked, this, [this](int fila, int) {
        QTableWidgetItem *item = ui->artistsTable->item(fila, 1);
        if (!item) {
            return;
        }
        int id = item->data(Qt::UserRole).toInt();
        QTableWidgetItem *tipoItem = ui->artistsTable->item(fila, 0);
        if (!tipoItem || tipoItem->text() != "Persona") {
            QMessageBox::information(
                this,
                "Consulta",
                "Solo consulta personas."
                );
            return;
        }
        consultaPersona(id);
    });
    connect(ui->addGroupButton, &QPushButton::clicked, this, &MainWindow::agregaGrupo);
    connect(ui->groupButton, &QPushButton::clicked, this, [this]() {
        ui->pagesStackedWidget->setCurrentWidget(ui->groupsPage);
        cargarGrupos();
    });
    connect(ui->groupsTable, &QTableWidget::cellDoubleClicked, this, [this](int fila, int) {
        QTableWidgetItem *item = ui->groupsTable->item(fila, 0);
        if (!item) {
            return;
        }
        int id = item->data(Qt::UserRole).toInt();
        consultaGrupo(id);
    });
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
    if (respuesta->url().path().startsWith("/api/grupos/") &&
        respuesta->operation() == QNetworkAccessManager::PutOperation) {
        QByteArray datos = respuesta->readAll();
        if (respuesta->error() != QNetworkReply::NoError) {
            QMessageBox::warning(
                this,
                "Error al editar",
                QString::fromUtf8(datos)
            );
        } else {
            QMessageBox::information(
                this,
                "Edición exitosa",
                "Los datos del grupo se actualizaron correctamente."
            );
            cargarGrupos();
        }
        respuesta->deleteLater();
        return;
    }
    if (respuesta->url().path().startsWith("/api/grupos/") &&
        respuesta->operation() == QNetworkAccessManager::GetOperation) {
        QByteArray datos = respuesta->readAll();
        if (respuesta->error() != QNetworkReply::NoError) {
            QMessageBox::warning(
                this,
                "Error",
                "No se pudieron consultar los datos del grupo:\n" +
                QString::fromUtf8(datos)
            );
            respuesta->deleteLater();
            return;
        }
        QJsonDocument documento = QJsonDocument::fromJson(datos);
        if (!documento.isObject()) {
            respuesta->deleteLater();
            return;
        }
        QJsonObject grupo = documento.object();
        int id = grupo["ID"].toInt();

        QDialog dialogo(this);
        dialogo.setWindowTitle("Datos del grupo");

        QFormLayout *formulario = new QFormLayout(&dialogo);
        QLineEdit *nombre = new QLineEdit(grupo["Nombre"].toString());
        QString fechaInicioOriginal = grupo["FechaInicio"].toString();
        QString fechaFinOriginal = grupo["FechaFin"].toString();

        QLineEdit *inicio = new QLineEdit(fechaInicioOriginal.isEmpty() ? "" : QDate::fromString(fechaInicioOriginal, Qt::ISODate).toString("dd/MM/yyyy"));

        QLineEdit *fin = new QLineEdit(fechaFinOriginal.isEmpty() ? "" : QDate::fromString(fechaFinOriginal, Qt::ISODate).toString("dd/MM/yyyy"));

        nombre->setReadOnly(true);
        inicio->setReadOnly(true);
        fin->setReadOnly(true);

        formulario->addRow("Nombre:", nombre);
        formulario->addRow("Fecha de inicio:", inicio);
        formulario->addRow("Fecha de fin:", fin);

        QPushButton *editar = new QPushButton("Editar");
        QPushButton *guardar = new QPushButton("Guardar");
        QPushButton *cerrar = new QPushButton("Cerrar");

        guardar->setEnabled(false);

        formulario->addRow(editar, guardar);
        formulario->addRow(cerrar);

        connect(editar, &QPushButton::clicked, &dialogo, [&]() {
            nombre->setReadOnly(false);
            inicio->setReadOnly(false);
            fin->setReadOnly(false);

            editar->setEnabled(false);
            guardar->setEnabled(true);
        });

        connect(cerrar, &QPushButton::clicked, &dialogo, &QDialog::reject);

        connect(guardar, &QPushButton::clicked, &dialogo, [&]() {
            QString nombreGrupo = nombre->text().trimmed();
            QString fechaInicio = inicio->text().trimmed();
            QString fechaFin = fin->text().trimmed();

            if (nombreGrupo.isEmpty()) {
                QMessageBox::warning(
                    &dialogo, "Error", "El nombre del grupo es obligatorio."
                    );
                return;
            }

            QDate inicioFecha = QDate::fromString(fechaInicio, "dd/MM/yyyy");
            QDate finFecha = QDate::fromString(fechaFin, "dd/MM/yyyy");

            if (!fechaInicio.isEmpty() && !inicioFecha.isValid()) {
                QMessageBox::warning(
                    &dialogo, "Error", "La fecha de inicio no es válida."
                    );
                return;
            }

            if (!fechaFin.isEmpty()) {
                if (!finFecha.isValid() || fechaInicio.isEmpty() ||
                    !inicioFecha.isValid() || finFecha <= inicioFecha) {
                    QMessageBox::warning(
                        &dialogo,
                        "Error",
                        "La fecha de fin debe ser posterior a la fecha de inicio."
                        );
                    return;
                }
            }

            QJsonObject datos;
            datos["Nombre"] = nombreGrupo;
            datos["FechaInicio"] = inicioFecha.isValid() ? inicioFecha.toString(Qt::ISODate) : "";
            datos["FechaFin"] = finFecha.isValid() ? finFecha.toString(Qt::ISODate) : "";

            QUrl direccion("http://127.0.0.1:8080/api/grupos/" + QString::number(id));

            QNetworkRequest peticion(direccion);
            peticion.setHeader(
                QNetworkRequest::ContentTypeHeader,
                "application/json"
            );

            QNetworkReply *respuesta = administradorRed->put(
                peticion,
                QJsonDocument(datos).toJson()
            );

            connect(respuesta, &QNetworkReply::finished, &dialogo, [&]() {
                if (respuesta->error() == QNetworkReply::NoError)
                    dialogo.accept();
            });
        });

        dialogo.exec();
        respuesta->deleteLater();
        return;
    }
    if (respuesta->url().path() == "/api/grupos" &&
        respuesta->operation() == QNetworkAccessManager::PostOperation) {
        QByteArray datos = respuesta->readAll();
        if (respuesta->error() != QNetworkReply::NoError) {
            QMessageBox::warning(
                this,
                "Error al crear grupo",
                QString::fromUtf8(datos)
            );
        } else {
            QMessageBox::information(
                this,
                "Registro correcto",
                "El grupo se agregó correctamente."
            );
            cargarGrupos();
        }
        respuesta->deleteLater();
        return;
    }
    if (respuesta->url().path().startsWith("/api/personas/") &&
        respuesta->operation() == QNetworkAccessManager::PutOperation) {

        QByteArray datos = respuesta->readAll();

        if (respuesta->error() != QNetworkReply::NoError) {
            QMessageBox::warning(this, "Error al editar", QString::fromUtf8(datos));
        } else {
            QMessageBox::information(this, "Edición exitosa", "Los datos de la persona se actualizaron correctamente.");
            QUrl direccion("http://127.0.0.1:8080/api/artistas");
            QNetworkRequest peticion(direccion);
            administradorRed->get(peticion);
        }
        respuesta->deleteLater();
        return;
    }
    if (respuesta->url().path().startsWith("/api/personas/") &&
        respuesta->operation() == QNetworkAccessManager::GetOperation) {

        QByteArray datos = respuesta->readAll();

        if (respuesta->error() != QNetworkReply::NoError) {
            QMessageBox::warning(this,
                "Error",
                "No se pudieron consultar los datos:\n" +
                QString::fromUtf8(datos)
            );
        } else {
            QJsonDocument documento = QJsonDocument::fromJson(datos);
            QJsonObject persona = documento.object();
            int id = respuesta->url().path().section('/', -1).toInt();

            QDialog dialogo(this);
            dialogo.setWindowTitle("Datos del artista");

            QFormLayout *formulario = new QFormLayout(&dialogo);

            QLineEdit *artistico = new QLineEdit(persona["NombreArtístico"].toString());
            QLineEdit *real = new QLineEdit(persona["NombreReal"].toString());
            QLineEdit *nacimiento = new QLineEdit(persona["FechaNacimiento"].toString());
            QLineEdit *fallecimiento = new QLineEdit(persona["FechaFallecimiento"].toString());

            artistico->setReadOnly(true);
            real->setReadOnly(true);
            nacimiento->setReadOnly(true);
            fallecimiento->setReadOnly(true);

            formulario->addRow("Nombre artístico:", artistico);
            formulario->addRow("Nombre real:", real);
            formulario->addRow("Fecha de nacimiento:", nacimiento);
            formulario->addRow("Fecha de fallecimiento:", fallecimiento);

            QDialogButtonBox *botones = new QDialogButtonBox(QDialogButtonBox::Close);

            QPushButton *editar = new QPushButton("Editar", &dialogo);

            botones->addButton(editar, QDialogButtonBox::ActionRole);

            QPushButton *guardar = new QPushButton("Guardar", &dialogo);

            botones->addButton(guardar, QDialogButtonBox::ActionRole);

            guardar->setEnabled(false);

            connect(editar, &QPushButton::clicked, &dialogo, [&]() {
                artistico->setReadOnly(false);
                real->setReadOnly(false);
                nacimiento->setReadOnly(false);
                fallecimiento->setReadOnly(false);

                editar->setEnabled(false);
                guardar->setEnabled(true);
            });

            connect(guardar, &QPushButton::clicked, &dialogo, [&, id]() {
                QString nombreArtistico = artistico->text().trimmed();
                QString nombreReal = real->text().trimmed();

                if (nombreArtistico.isEmpty() && nombreReal.isEmpty()) {
                    QMessageBox::warning(
                        &dialogo,
                        "Datos incompletos",
                        "Debes proporcionar al menos un nombre."
                    );
                    return;
                }
                QString nacimientoTexto = nacimiento->text().trimmed();
                QString fallecimientoTexto = fallecimiento->text().trimmed();

                QDate fechaNac = QDate::fromString(nacimientoTexto, Qt::ISODate);
                QDate fechaFall = QDate::fromString(fallecimientoTexto, Qt::ISODate);

                if (!nacimientoTexto.isEmpty() && !fechaNac.isValid()) {
                    QMessageBox::warning(&dialogo, "Fecha inválida", "La fecha de nacimiento es inválida.");
                    return;
                }

                if (!fallecimientoTexto.isEmpty() && !fechaFall.isValid()) {
                    QMessageBox::warning(&dialogo, "Fecha inválida", "La fecha de fallecimiento es inválida.");
                    return;
                }

                if (!fallecimientoTexto.isEmpty() &&
                    (nacimientoTexto.isEmpty() || fechaFall <= fechaNac)) {
                    QMessageBox::warning(&dialogo, "Fechas inválidas",
                    "La fecha de fallecimiento debe ser posterior al nacimiento.");
                    return;
                }
                QJsonObject solicitud;
                solicitud["NombreArtístico"] = nombreArtistico;
                solicitud["NombreReal"] = nombreReal;
                solicitud["FechaNacimiento"] = nacimientoTexto;
                solicitud["FechaFallecimiento"] = fallecimientoTexto;

                QUrl direccion(
                    "http://127.0.0.1:8080/api/personas/" +
                    QString::number(id)
                );

                QNetworkRequest peticion(direccion);
                peticion.setHeader(
                    QNetworkRequest::ContentTypeHeader,
                    "application/json"
                );
                QNetworkReply *respuesta = administradorRed->put(peticion,
                    QJsonDocument(solicitud).toJson());

                connect(respuesta, &QNetworkReply::finished, &dialogo, [&]() {
                    if (respuesta->error() == QNetworkReply::NoError) {
                        dialogo.accept();
                    }
                });
            });

            connect(botones, &QDialogButtonBox::rejected, &dialogo, &QDialog::reject);

            formulario->addRow(botones);

            dialogo.exec();
        }
        respuesta->deleteLater();
        return;
    }

    if (respuesta->url().path() == "/api/personas" &&
        respuesta->operation() == QNetworkAccessManager::PostOperation) {

        if (respuesta->error() != QNetworkReply::NoError) {
            QMessageBox::warning(
                this,
                "Error",
                "No se pudo registrar al artista:\n" +
                    respuesta->readAll()
            );
        } else {
            QMessageBox::information(
                this,
                "Registro exitoso",
                "El artista se agregó correctamente."
            );

            QUrl direccion("http://127.0.0.1:8080/api/artistas");
            QNetworkRequest peticion(direccion);
            administradorRed->get(peticion);
        }
        respuesta->deleteLater();
        return;
    }

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
    if (respuesta->url().path() == "/api/grupos") {
        QJsonDocument documento = QJsonDocument::fromJson(datos);

        if (!documento.isArray()) {
            respuesta->deleteLater();
            return;
        }

        QJsonArray grupos = documento.array();

        ui->groupsTable->setRowCount(0);
        for (const QJsonValue &valor : grupos) {
            QJsonObject grupo = valor.toObject();

            int fila = ui->groupsTable->rowCount();
            ui->groupsTable->insertRow(fila);
            QTableWidgetItem *nombreItem = new QTableWidgetItem(grupo["Nombre"].toString());

            nombreItem->setData(Qt::UserRole, grupo["ID"].toInt());

            ui->groupsTable->setItem(fila, 0, nombreItem);

            ui->groupsTable->setItem(
                fila, 1,
                new QTableWidgetItem(grupo["FechaInicio"].toString())
            );

            ui->groupsTable->setItem(
                fila, 2,
                new QTableWidgetItem(grupo["FechaFin"].toString())
            );
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

    QString nomArt = nombreArtistico->text().trimmed();
    QString nomReal = nombreReal->text().trimmed();

    if (nomArt.isEmpty() && nomReal.isEmpty())
    {
        QMessageBox::warning(
            this,
            "Datos incompletos",
            "Debes ingresar al menos un nombre artístico o real."
        );
        return;
    }

    if (tieneNacimiento->isChecked() &&
        tieneFallecimiento->isChecked()) {

        if (fechaNacimiento->date() >= fechaFallecimiento->date()) {
            QMessageBox::warning(this,
                "Fechas inválidas", "La fecha de fallecimiento debe ser posterior a la fecha de nacimiento."
                );

            return;
        }
    }

    QJsonObject solicitud;

    solicitud["NombreArtístico"] = nomArt;
    solicitud["NombreReal"] = nomReal;

    solicitud["FechaNacimiento"] = tieneNacimiento->isChecked() ? fechaNacimiento->date().toString("yyyy-MM-dd") : "";

    solicitud["FechaFallecimiento"] = tieneFallecimiento->isChecked() ? fechaFallecimiento->date().toString("yyyy-MM-dd") : "";

    QUrl direccion("http://127.0.0.1:8080/api/personas");

    QNetworkRequest peticion(direccion);

    peticion.setHeader(
        QNetworkRequest::ContentTypeHeader,
        "application/json"
    );

    administradorRed->post(
        peticion,
        QJsonDocument(solicitud).toJson()
    );
}

void MainWindow::consultaPersona(int id)
{
    QUrl direccion(
        "http://127.0.0.1:8080/api/personas/" +
        QString::number(id)
    );

    QNetworkRequest peticion(direccion);
    administradorRed->get(peticion);
}

void MainWindow::consultaGrupo(int id)
{
    QUrl direccion(
        "http://127.0.0.1:8080/api/grupos/" +
        QString::number(id)
    );
    QNetworkRequest peticion(direccion);
    administradorRed->get(peticion);
}

void MainWindow::agregaGrupo()
{
    QDialog dialogo(this);
    dialogo.setWindowTitle("Crear grupo");

    QFormLayout *formulario = new QFormLayout(&dialogo);
    QLineEdit *nombre = new QLineEdit(&dialogo);

    QCheckBox *activarInicio = new QCheckBox("Registrar fecha de inicio");
    QDateEdit *fechaInicio = new QDateEdit(QDate::currentDate());
    fechaInicio->setDisplayFormat("dd/MM/yyyy");
    fechaInicio->setCalendarPopup(true);
    fechaInicio->setEnabled(false);

    QCheckBox *activarFin = new QCheckBox("Registrar fecha de fin");
    QDateEdit *fechaFin = new QDateEdit(QDate::currentDate());
    fechaFin->setDisplayFormat("dd/MM/yyyy");
    fechaFin->setCalendarPopup(true);
    fechaFin->setEnabled(false);

    connect(activarInicio, &QCheckBox::toggled, fechaInicio, &QDateEdit::setEnabled);
    connect(activarFin, &QCheckBox::toggled, fechaFin, &QDateEdit::setEnabled);

    activarFin->setEnabled(false);

    connect(activarInicio, &QCheckBox::toggled, activarFin, &QCheckBox::setEnabled);

    connect(activarInicio, &QCheckBox::toggled, [=](bool activado) {
        if (!activado)
            activarFin->setChecked(false);
    });

    formulario->addRow("Nombre:", nombre);
    formulario->addRow(activarInicio);
    formulario->addRow("Fecha de inicio:", fechaInicio);
    formulario->addRow(activarFin);
    formulario->addRow("Fecha de fin:", fechaFin);

    QDialogButtonBox *botones = new QDialogButtonBox(QDialogButtonBox::Ok | QDialogButtonBox::Cancel);

    botones->button(QDialogButtonBox::Ok)->setText("Crear");
    botones->button(QDialogButtonBox::Cancel)->setText("Cancelar");

    formulario->addRow(botones);

    connect(botones, &QDialogButtonBox::accepted, [&]() {
        if (nombre->text().trimmed().isEmpty()) {
            QMessageBox::warning(
                &dialogo,
                "Datos incompletos",
                "El nombre del grupo es obligatorio."
            );
            return;
        }
        if (activarFin->isChecked() &&
            fechaFin->date() <= fechaInicio->date()) {
            QMessageBox::warning(
                &dialogo,
                "Fechas incorrectas",
                "La fecha de fin debe ser posterior a la fecha de inicio."
                );
            return;
        }
        QJsonObject solicitud;
        solicitud["Nombre"] = nombre->text().trimmed();
        solicitud["FechaInicio"] = activarInicio->isChecked() ? fechaInicio->date().toString(Qt::ISODate) : "";
        solicitud["FechaFin"] = activarFin->isChecked() ? fechaFin->date().toString(Qt::ISODate) : "";
        QUrl direccion("http://127.0.0.1:8080/api/grupos");
        QNetworkRequest peticion(direccion);
        peticion.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");
        QNetworkReply *respuesta = administradorRed->post(peticion, QJsonDocument(solicitud).toJson());
        dialogo.accept();
    });
    connect(botones, &QDialogButtonBox::rejected, &dialogo, &QDialog::reject);

    dialogo.exec();
}

void MainWindow::cargarGrupos()
{
    QUrl direccion("http://127.0.0.1:8080/api/grupos");
    QNetworkRequest peticion(direccion);
    administradorRed->get(peticion);
}
