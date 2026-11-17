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
#include <QSpinBox>
#include <QComboBox>

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , ui(new Ui::MainWindow)
    , administradorRed(new QNetworkAccessManager(this))
    , reproductor(new QMediaPlayer(this))
    , salidaAudio(new QAudioOutput(this))
{
    ui->setupUi(this);
    connect(ui->songsTable, &QTableWidget::cellDoubleClicked, this, &MainWindow::editaRola);
    reproductor->setAudioOutput(salidaAudio);
    connect(administradorRed, &QNetworkAccessManager::finished, this, &MainWindow::procesaRes);

    QUrl direccionArtistas("http://127.0.0.1:8080/api/artistas");
    QNetworkRequest peticionArtistas(direccionArtistas);
    administradorRed->get(peticionArtistas);

    QUrl direccionAlbumes("http://127.0.0.1:8080/api/albumes");
    QNetworkRequest peticionAlbumes(direccionAlbumes);
    administradorRed->get(peticionAlbumes);

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
    connect(ui->albumsListWidget, &QListWidget::itemDoubleClicked, this, [this](QListWidgetItem *elemento) {
        int id = elemento->data(Qt::UserRole).toInt();

        if (id <= 0) {
            return;
        }

        QUrl direccion(
            "http://127.0.0.1:8080/api/albumes/" +
            QString::number(id)
        );

        QNetworkRequest peticion(direccion);
        administradorRed->get(peticion);
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
        if (!tipoItem) {
            return;
        }
        if (tipoItem->text() == "Persona") {
            consultaPersona(id);
        } else if (tipoItem->text() == "Grupo") {
            consultaGrupo(id);
        } else if (tipoItem->text() == "Desconocido") {
            QMessageBox pregunta(this);
            pregunta.setWindowTitle("Intérprete desconocido");
            pregunta.setText("¿Qué tipo de intérprete deseas registrar?");

            QPushButton *artista = pregunta.addButton("Artista", QMessageBox::AcceptRole);
            QPushButton *grupo = pregunta.addButton("Grupo", QMessageBox::AcceptRole);

            pregunta.addButton("Cancelar", QMessageBox::RejectRole);
            pregunta.exec();

            if (pregunta.clickedButton() == artista) {
                agregaPersona();
            } else if (pregunta.clickedButton() == grupo) {
                agregaGrupo();
            }
        }

    });
    /* connect(ui->songsTable, &QTableWidget::cellDoubleClicked, this, [this](int fila, int columna) {
        QTableWidgetItem *tituloItem = ui->songsTable->item(fila, 0);
        QTableWidgetItem *artistaItem = ui->songsTable->item(fila, 1);

        if (!tituloItem || !artistaItem)
            return;

        if (artistaItem->text() != "Desconocido")
            return;

        int idRola = tituloItem->data(Qt::UserRole + 1).toInt();
        if (idRola <= 0)
            return;

        idRolaPendiente = idRola;

        QMessageBox pregunta(this);
        pregunta.setWindowTitle("Corregir intérprete");
        pregunta.setText("¿Qué tipo de intérprete deseas registrar?");

        QPushButton *artista = pregunta.addButton("Artista", QMessageBox::AcceptRole);
        QPushButton *grupo = pregunta.addButton("Grupo", QMessageBox::AcceptRole);

        pregunta.addButton("Cancelar", QMessageBox::RejectRole);

        pregunta.exec();

        if (pregunta.clickedButton() == artista) {
            agregaPersona();
        }
        else if (pregunta.clickedButton() == grupo) {
            agregaGrupo();
        }
        else {
            idRolaPendiente = 0;
        }
    });*/
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
    ui->songsTable->setEditTriggers(QAbstractItemView::NoEditTriggers);
    ui->artistsTable->setEditTriggers(QAbstractItemView::NoEditTriggers);
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
        respuesta->url().path().contains("/integrantes")) {
        return;
    }
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
        QListWidget *integrantes = new QListWidget();
        formulario->addRow("Integrantes:", integrantes);
        cargarIntegrantes(id, integrantes);

        QPushButton *agregar = new QPushButton("Agregar integrante");
        formulario->addRow(agregar);

        connect(agregar, &QPushButton::clicked, &dialogo, [this, id, integrantes, &dialogo]() {
            QDialog seleccion(&dialogo);
            seleccion.setWindowTitle("Agregar integrante");

            QVBoxLayout *layout = new QVBoxLayout(&seleccion);

            QListWidget *personas = new QListWidget();
            layout->addWidget(personas);

            QPushButton *cerrar = new QPushButton("Cerrar");
            layout->addWidget(cerrar);

            QPushButton *confirmar = new QPushButton("Agregar seleccionado");
            layout->addWidget(confirmar);

            connect(cerrar, &QPushButton::clicked, &seleccion, &QDialog::reject);
            QUrl direccionPersonas("http://127.0.0.1:8080/api/personas");

            QNetworkRequest peticionPersonas(direccionPersonas);
            QNetworkReply *respuestaPersonas = administradorRed->get(peticionPersonas);

            connect(respuestaPersonas, &QNetworkReply::finished, &seleccion, [respuestaPersonas, personas]() {
                if (respuestaPersonas->error() != QNetworkReply::NoError) {
                    QMessageBox::warning(
                        nullptr,
                        "Error",
                        "No se pudieron consultar las personas."
                    );
                    respuestaPersonas->deleteLater();
                    return;
                }

                QJsonDocument documento = QJsonDocument::fromJson(respuestaPersonas->readAll());

                if (documento.isArray()) {
                    for (const QJsonValue &valor : documento.array()) {
                        QJsonObject persona = valor.toObject();

                        QString nombre = persona["NombreArtístico"].toString();

                        if (nombre.isEmpty())
                            nombre = persona["NombreReal"].toString();

                        QListWidgetItem *elemento = new QListWidgetItem(nombre);

                        elemento->setData(Qt::UserRole, persona["ID"].toInt());

                        personas->addItem(elemento);
                    }
                }
                respuestaPersonas->deleteLater();
            });

            connect(confirmar, &QPushButton::clicked, &seleccion, [this, id, integrantes, personas, &seleccion]() {
                QListWidgetItem *elemento = personas->currentItem();
                if (!elemento) {
                    QMessageBox::warning(
                    &seleccion,
                    "Atención",
                    "Selecciona una persona."
                    );
                    return;
                }

                int idPersona = elemento->data(Qt::UserRole).toInt();

                QUrl direccion(
                    "http://127.0.0.1:8080/api/grupos/" +
                    QString::number(id) + "/integrantes"
                );

                QNetworkRequest peticion(direccion);

                peticion.setHeader(
                    QNetworkRequest::ContentTypeHeader,
                    "application/json"
                );

                QJsonObject datos;
                datos["idPersona"] = idPersona;

                QNetworkReply *respuesta = administradorRed->post(
                    peticion,
                    QJsonDocument(datos).toJson()
                );

                connect(respuesta, &QNetworkReply::finished, &seleccion, [respuesta, id, integrantes, &seleccion, this]() {
                    QByteArray respuestaDatos = respuesta->readAll();
                    if (respuesta->error() != QNetworkReply::NoError) {
                        QMessageBox::warning(
                        &seleccion,
                        "Error",
                        QString::fromUtf8(respuestaDatos)
                        );
                    } else {
                        QMessageBox::information(
                        &seleccion,
                        "Éxito",
                        "Integrante agregado correctamente."
                        );

                        cargarIntegrantes(id, integrantes);
                        seleccion.accept();
                    }

                    respuesta->deleteLater();
                });
            });
            seleccion.exec();
        });
        QPushButton *eliminar = new QPushButton("Eliminar integrante");
        formulario->addRow(eliminar);
        connect(eliminar, &QPushButton::clicked, &dialogo, [this, id, integrantes, &dialogo]() {
            QListWidgetItem *elemento = integrantes->currentItem();
            if (!elemento) {
                QMessageBox::warning(
                    &dialogo,
                    "Atención",
                    "Selecciona un integrante para eliminar."
                );
                return;
            }

            int idPersona = elemento->data(Qt::UserRole).toInt();

            auto confirmacion = QMessageBox::question(
                &dialogo,
                "Confirmar eliminación",
                "¿Quieres eliminar a este integrante del grupo?"
            );

            if (confirmacion != QMessageBox::Yes)
                return;

            QUrl direccion(
                "http://127.0.0.1:8080/api/grupos/" +
                QString::number(id) + "/integrantes/" +
                QString::number(idPersona)
            );

            QNetworkRequest peticion(direccion);
            QNetworkReply *respuesta = administradorRed->deleteResource(peticion);

            connect(respuesta, &QNetworkReply::finished, &dialogo, [this, respuesta, id, integrantes, &dialogo]() {
                QByteArray datos = respuesta->readAll();
                if (respuesta->error() != QNetworkReply::NoError) {
                    QMessageBox::warning(
                        &dialogo,
                        "Error",
                        QString::fromUtf8(datos)
                    );
                } else {
                    QMessageBox::information(
                        &dialogo,
                        "Éxito",
                        "Integrante eliminado correctamente."
                    );

                    cargarIntegrantes(id, integrantes);
                }

                respuesta->deleteLater();
            });
        });
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

    if (respuesta->url().path().startsWith("/api/rolas/") &&
        respuesta->url().path().endsWith("/artista") &&
        respuesta->operation() == QNetworkAccessManager::PutOperation) {

        QByteArray datos = respuesta->readAll();

        if (respuesta->error() != QNetworkReply::NoError) {
            QMessageBox::warning(
                this,
                "Error al corregir intérprete",
                QString::fromUtf8(datos)
            );
        } else {
            QMessageBox::information(
                this,
                "Éxito",
                "El intérprete de la canción se actualizó correctamente."
            );

            idRolaPendiente = 0;

            QUrl direccionRolas("http://127.0.0.1:8080/api/rolas");
            QNetworkRequest peticionRolas(direccionRolas);
            administradorRed->get(peticionRolas);

            QUrl direccionArtistas("http://127.0.0.1:8080/api/artistas");
            QNetworkRequest peticionArtistas(direccionArtistas);
            administradorRed->get(peticionArtistas);
        }
        respuesta->deleteLater();
        return;
    }

    if (respuesta->url().path() == "/api/personas" &&
        respuesta->operation() == QNetworkAccessManager::PostOperation) {

        QByteArray datos = respuesta->readAll();

        if (respuesta->error() != QNetworkReply::NoError) {
            QMessageBox::warning(
                this,
                "Error",
                "No se pudo registrar al artista:\n" +
                    QString::fromUtf8(datos)
            );
        } else {
            QJsonDocument documento = QJsonDocument::fromJson(datos);
            int idPerformer = documento.object()["id"].toInt();

            if (idPerformer <= 0) {
                QMessageBox::warning(
                    this,
                    "Error",
                    "No se recibió un Id válido del artista."
                    );
                respuesta->deleteLater();
                return;
            }
            QMessageBox::information(
                this,
                "Registro exitoso",
                "El artista se agregó correctamente."
            );

            if (idRolaPendiente > 0) {
                QUrl direccion(
                    "http://127.0.0.1:8080/api/rolas/" +
                    QString::number(idRolaPendiente) + "/artista"
                );

                QNetworkRequest peticion(direccion);

                peticion.setHeader(
                    QNetworkRequest::ContentTypeHeader,
                    "application/json"
                );

                QJsonObject solicitud;
                solicitud["idPerformer"] = idPerformer;

                administradorRed->put(
                    peticion,
                    QJsonDocument(solicitud).toJson()
                );
            } else {
                QUrl direccion("http://127.0.0.1:8080/api/artistas");
                QNetworkRequest peticion(direccion);
                administradorRed->get(peticion);
            }
        }
        respuesta->deleteLater();
        return;
    }

    if (respuesta->url().path() == "/api/personas" &&
        respuesta->operation() == QNetworkAccessManager::GetOperation) {
        respuesta->deleteLater();
        return;
    }
    QByteArray datos = respuesta->readAll();

    if (respuesta->url().path().startsWith("/api/rolas/") &&
        respuesta->operation() == QNetworkAccessManager::PutOperation) {

        if (respuesta->error() == QNetworkReply::NoError) {
            QMessageBox::information(
                this,
                "Éxito",
                "La rola se actualizó correctamente."
            );
            cargarRolas();
        } else {
            QMessageBox::critical(
                this,
                "Error",
                "No se pudo actualizar la rola."
            );
        }

        respuesta->deleteLater();
        return;
    }

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
        artistasDisponibles = artistas;

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
    if (respuesta->url().path() == "/api/albumes" &&
        respuesta->operation() == QNetworkAccessManager::GetOperation) {

        QJsonDocument documento = QJsonDocument::fromJson(datos);
        if (!documento.isArray()) {
            respuesta->deleteLater();
            return;
        }

        albumesDisponibles = documento.array();

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

    if (respuesta->url().path().startsWith("/api/albumes/") &&
        respuesta->operation() == QNetworkAccessManager::PutOperation) {

        if (respuesta->error() != QNetworkReply::NoError) {
            QMessageBox::critical(
                this,
                "Error",
                "No se pudo actualizar el álbum:\n" +
                    respuesta->errorString()
            );
        } else {
            QMessageBox::information(
                this,
                "Éxito",
                "Los datos del álbum se actualizaron correctamente."
            );
            cargarRolas();

            QUrl direccion("http://127.0.0.1:8080/api/albumes");
            QNetworkRequest peticion(direccion);
            administradorRed->get(peticion);
        }

        respuesta->deleteLater();
        return;
    }

    if (respuesta->url().path().startsWith("/api/albumes/") &&
        respuesta->operation() == QNetworkAccessManager::GetOperation) {

        QJsonDocument documento = QJsonDocument::fromJson(datos);

        if (!documento.isObject()) {
            respuesta->deleteLater();
            return;
        }

        QJsonObject album = documento.object();

        QDialog dialogo(this);
        dialogo.setWindowTitle("Datos del álbum");

        QFormLayout formulario(&dialogo);

        QLineEdit nombre;
        QSpinBox año;
        QLineEdit ruta;

        QDialogButtonBox botones(QDialogButtonBox::Close);

        QPushButton editar("Editar");
        QPushButton guardar("Guardar");

        botones.addButton(&editar, QDialogButtonBox::ActionRole);
        botones.addButton(&guardar, QDialogButtonBox::ActionRole);

        guardar.setEnabled(false);

        formulario.addRow(&botones);

        connect(&editar, &QPushButton::clicked, &dialogo, [&]() {
            nombre.setEnabled(true);
            año.setEnabled(true);
            editar.setEnabled(false);
            guardar.setEnabled(true);
        });

        connect(&guardar, &QPushButton::clicked, &dialogo, [&]() {
            QJsonObject cambios;

            cambios["nombre"] = nombre.text();
            cambios["año"] = año.value();
            cambios["ruta"] = ruta.text();

            QJsonDocument documentoEnvio(cambios);
            QUrl direccion(
                "http://127.0.0.1:8080/api/albumes/" +
                QString::number(album["id"].toInt())
            );

            QNetworkRequest peticion(direccion);
            peticion.setHeader(
                QNetworkRequest::ContentTypeHeader,
                "application/json"
            );

            administradorRed->put(
                peticion,
                documentoEnvio.toJson()
            );
            dialogo.accept();
        });

        connect(&botones, &QDialogButtonBox::rejected, &dialogo, &QDialog::reject);

        nombre.setText(album["nombre"].toString());
        año.setRange(0, 3000);
        año.setValue(album["año"].toInt());
        ruta.setText(album["ruta"].toString());

        ruta.setReadOnly(true);
        nombre.setEnabled(false);
        año.setEnabled(false);

        formulario.addRow("Nombre:", &nombre);
        formulario.addRow("Año:", &año);
        formulario.addRow("Directorio:", &ruta);
        dialogo.exec();
        respuesta->deleteLater();
        return;
    }

    if (respuesta->url().path().startsWith("/api/rolas/") &&
        respuesta->operation() == QNetworkAccessManager::GetOperation) {

        QJsonDocument documento = QJsonDocument::fromJson(datos);

        if (!documento.isObject()) {
            respuesta->deleteLater();
            return;
        }

        QJsonObject rola = documento.object();

        QDialog dialogo(this);
        dialogo.setWindowTitle("Editar rola");

        QFormLayout formulario(&dialogo);

        QLineEdit titulo;
        QComboBox artista;
        QComboBox album;
        QSpinBox pista;
        QSpinBox año;
        QLineEdit genero;

        album.setEditable(false);
        artista.setEditable(false);

        titulo.setText(rola["title"].toString());

        artista.clear();

        int indiceActual = -1;

        for (const QJsonValue &valor : artistasDisponibles) {
            QJsonObject datoArtista = valor.toObject();

            QString nombre = datoArtista["Nombre"].toString();
            int id = datoArtista["Id"].toInt();

            artista.addItem(nombre, id);

            if (id == rola["performer_id"].toInt()) {
                indiceActual = artista.count() - 1;
            }
        }

        if (indiceActual >= 0) {
            artista.setCurrentIndex(indiceActual);
        } else if (rola["performer_id"].toInt() > 0) {
            artista.addItem(
                rola["artist"].toString(),
                rola["performer_id"].toInt()
            );

            artista.setCurrentIndex(artista.count() - 1);
        }

        pista.setRange(0, 999);
        año.setRange(0, 3000);

        album.clear();

        int indiceAlbumActual = -1;

        for (const QJsonValue &valor : albumesDisponibles) {
            QJsonObject datoAlbum = valor.toObject();

            QString nombre = datoAlbum["nombre"].toString();
            int id = datoAlbum["id"].toInt();

            album.addItem(nombre, id);

            if (id == rola["album_id"].toInt()) {
                indiceAlbumActual = album.count() - 1;
            }
        }

        if (indiceAlbumActual >= 0) {
            album.setCurrentIndex(indiceAlbumActual);
        } else if (rola["album_id"].toInt() > 0) {
            album.addItem(
                rola["album"].toString(),
                rola["album_id"].toInt()
            );

            album.setCurrentIndex(album.count() - 1);
        }

        pista.setValue(rola["track"].toInt());
        año.setValue(rola["year"].toInt());
        genero.setText(rola["genre"].toString());

        formulario.addRow("Título:", &titulo);
        formulario.addRow("Artista:", &artista);
        formulario.addRow("Álbum:", &album);
        formulario.addRow("Número de pista:", &pista);
        formulario.addRow("Año:", &año);
        formulario.addRow("Género:", &genero);

        QDialogButtonBox botones(QDialogButtonBox::Save | QDialogButtonBox::Cancel);

        formulario.addRow(&botones);
        QObject::connect(&botones, &QDialogButtonBox::accepted, &dialogo, &QDialog::accept);
        QObject::connect(&botones, &QDialogButtonBox::rejected, &dialogo, &QDialog::reject);

        if (dialogo.exec() != QDialog::Accepted) {
            respuesta->deleteLater();
            return;
        }

        QJsonObject cambios;

        cambios["title"] = titulo.text();
        cambios["performer_id"] = artista.currentData().toInt();
        cambios["album_id"] = album.currentData().toInt();
        cambios["path"] = rola["path"].toString();
        cambios["track"] = pista.value();
        cambios["year"] = año.value();
        cambios["genre"] = genero.text();

        QJsonDocument documentoEnvio(cambios);
        QByteArray datosEnvio = documentoEnvio.toJson();

        QUrl direccion("http://127.0.0.1:8080/api/rolas/" +
                       QString::number(rola["id"].toInt()));

        QNetworkRequest peticion(direccion);
        peticion.setHeader(QNetworkRequest::ContentTypeHeader, "application/json");

        administradorRed->put(peticion, datosEnvio);

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
        tituloItem->setData(Qt::UserRole + 1, rola["id"].toInt());
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

void MainWindow::cargarIntegrantes(int idGrupo, QListWidget *lista)
{
    QUrl direccion(
        "http://127.0.0.1:8080/api/grupos/" +
        QString::number(idGrupo) + "/integrantes"
    );
    QNetworkRequest peticion(direccion);
    QNetworkReply *respuesta = administradorRed->get(peticion);

    connect(respuesta, &QNetworkReply::finished, lista, [respuesta, lista]() {
        if (respuesta->error() != QNetworkReply::NoError) {
            QMessageBox::warning(
                nullptr,
                "Error",
                "No se pudieron consultar los integrantes."
            );
            respuesta->deleteLater();
            return;
        }

        QJsonDocument documento = QJsonDocument::fromJson(respuesta->readAll());

        if (documento.isArray() && lista) {
            lista->clear();
            for (const QJsonValue &valor : documento.array()) {
                QJsonObject persona = valor.toObject();
                QString nombre = persona["NombreArtístico"].toString();

                if (nombre.isEmpty())
                    nombre = persona["NombreReal"].toString();

                QListWidgetItem *elemento = new QListWidgetItem(nombre);

                elemento->setData(Qt::UserRole, persona["ID"].toInt());

                lista->addItem(elemento);
            }
        }

        respuesta->deleteLater();
    });
}

void MainWindow::editaRola(int fila, int columna)
{
    Q_UNUSED(columna);
    QTableWidgetItem *tituloItem = ui->songsTable->item(fila, 0);

    if (!tituloItem) {
        return;
    }

    int id = tituloItem->data(Qt::UserRole + 1).toInt();

    if (id <= 0) {
        return;
    }

    consultaRola(id);
}

void MainWindow::consultaRola(int id)
{
    QUrl direccion("http://127.0.0.1:8080/api/rolas/" + QString::number(id));
    QNetworkRequest peticion(direccion);
    administradorRed->get(peticion);
}
