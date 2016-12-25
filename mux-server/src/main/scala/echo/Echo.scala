package echo

import com.twitter.finagle._
import com.twitter.util._

object Main {

  val svc = Service.mk { req: mux.Request =>
    Future.value(mux.Response(req.body))
  }

  def main(args: Array[String]): Unit = {
    println("Hello, world!")

    val server = Mux.server.serve(":8081", svc)
    Await.ready(server)
  }
}

