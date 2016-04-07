
void display(bool showChannelIndex = false, TString fileNameCenters="../geom/dpgageom.csv", TString fileNameFull="../geom/dpgafullgeom.csv", Bool_t register=kTRUE)
{
  TEveManager::Create();

  gStyle->SetPalette(1, 0);
  
  TEveRGBAPalette* pal = new TEveRGBAPalette(0, 130);
  
  // Draw coordinate system
  TEveArrow* xAxis = new TEveArrow(100., 0., 0., 0., 0., 0.);
  TEveArrow* yAxis = new TEveArrow(0., 100., 0., 0., 0., 0.);
  TEveArrow* zAxis = new TEveArrow(0., 0., 100., 0., 0., 0.);
  xAxis->SetMainColor(kBlue); xAxis->SetTubeR(0.02); xAxis->SetPickable(kTRUE);
  yAxis->SetMainColor(kBlue); yAxis->SetTubeR(0.02); yAxis->SetPickable(kTRUE);
  zAxis->SetMainColor(kBlue); zAxis->SetTubeR(0.02); zAxis->SetPickable(kTRUE);
  gEve->AddElement(xAxis);
  gEve->AddElement(yAxis);
  gEve->AddElement(zAxis);
  TEveText* tx = new TEveText("x"); tx->SetFontSize(20);
  TEveVector tvx = xAxis->GetVector()*1.1+xAxis->GetOrigin(); tx->RefMainTrans().SetPos(tvx.Arr());
  xAxis->AddElement(tx);
  TEveText* ty = new TEveText("y"); ty->SetFontSize(20);
  TEveVector tvy = yAxis->GetVector()*1.1+yAxis->GetOrigin(); ty->RefMainTrans().SetPos(tvy.Arr());
  yAxis->AddElement(ty);
  TEveText* tz = new TEveText("z"); tz->SetFontSize(20);
  TEveVector tvz = zAxis->GetVector()*1.1+zAxis->GetOrigin(); tz->RefMainTrans().SetPos(tvz.Arr());
  zAxis->AddElement(tz);
  
  // Read full coordinates of scintillators and put them in a box set
  TEveBoxSet* bsright = new TEveBoxSet("BoxSetRight");
  bsright->Reset(TEveBoxSet::kBT_FreeBox, kFALSE, 64);
  bsright->SetPalette(pal);

  TEveBoxSet* bsleft = new TEveBoxSet("BoxSetLeft");
  bsleft->Reset(TEveBoxSet::kBT_FreeBox, kFALSE, 64);
  bsleft->SetPalette(pal);

  ifstream infull(fileNameFull.Data());
  if (!infull) {
    cerr << "ERROR ! Unable to open file '" << fileNameFull << "' !" << endl;
    return;
  }
  for(; !infull.eof() ;) {
    string line;
    if (!getline(infull,line)) break;
    if (!line.empty() && line[0]!='#') {
      istringstream istr(line);
      int iChannelAbs240;
      float X0, Y0, Z0;
      float X1, Y1, Z1;
      float X2, Y2, Z2;
      float X3, Y3, Z3;
      float X4, Y4, Z4;
      float X5, Y5, Z5;
      float X6, Y6, Z6;
      float X7, Y7, Z7;

      istr >> iChannelAbs240
	   >> X0 >> Y0 >> Z0
	   >> X1 >> Y1 >> Z1
	   >> X2 >> Y2 >> Z2
	   >> X3 >> Y3 >> Z3
	   >> X4 >> Y4 >> Z4
	   >> X5 >> Y5 >> Z5
	   >> X6 >> Y6 >> Z6
	   >> X7 >> Y7 >> Z7;
       
      Float_t verts[24] = {
	X0 , Y0 , Z0, 
	X1 , Y1 , Z1,
	X2 , Y2 , Z2,
	X3 , Y3 , Z3,
	X4 , Y4 , Z4,
	X5 , Y5 , Z5,
	X6 , Y6 , Z6,
	X7 , Y7 , Z7};
      
      // Color code:
      //  100 -> yellow
      //  10 -> blue
      if(iChannelAbs240 < 120) {
	bsright->AddBox(verts);
	bsright->DigitValue(100);
      }
      else {
	bsleft->AddBox(verts);
	bsleft->DigitValue(10);
      }
    }
  }
  bsright->RefitPlex();
  bsleft->RefitPlex();

  // Read centers of scintillators' front faces and put them in a point set
  TEvePointSet* ps = new TEvePointSet();
  ps->SetOwnIds(kTRUE);
  
  ifstream incenters(fileNameCenters.Data());
  if (!incenters) {
    cerr << "ERROR ! Unable to open file '" << fileNameCenters << "' !" << endl;
    return;
  }
  for(; !incenters.eof() ;) {
    string line;
    if (!getline(incenters,line)) break;
    if (!line.empty() && line[0]!='#') {
      istringstream istr(line);
      int iChannelAbs240;
      float X, Y, Z;

      istr >> iChannelAbs240
	   >> X >> Y >> Z;
       
      ps -> SetNextPoint(X, Y, Z);
      ps -> SetPointId(new TNamed(Form("iChannelAbs240=%d", iChannelAbs240), ""));

      if(showChannelIndex) {
	float x, y, z;
	ps -> GetPoint(ps->GetLastPoint(), x, y, z);
	TEveText* text = new TEveText(Form("%i", iChannelAbs240)); text->SetFontSize(10);
	if (iChannelAbs240 >= 120) {
	  text->RefMainTrans().SetPos(x-3, y-3, z);
      }
	else {
	  text->RefMainTrans().SetPos(x+3, y+3, z);
	}
	gEve->AddElement(text);
      }
    }
  }

  ps->SetMarkerColor(kRed);
  ps->SetMarkerSize(1);
  ps->SetMarkerStyle(8);

  if (register)
    {
      gEve->AddElement(bsright);
      gEve->AddElement(bsleft);
      gEve->AddElement(ps);
      gEve->Redraw3D(kTRUE);
    }
}
